// internal/module/account/google_oauth/service.go
package google_oauth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"

	"postmatic-api/config"
	"postmatic-api/internal/module/headless/mailer"
	"postmatic-api/internal/module/headless/queue"
	"postmatic-api/internal/module/headless/token"
	"postmatic-api/internal/repository/entity"
	emailLimiterRepo "postmatic-api/internal/repository/redis/email_limiter_repository"
	sessRepo "postmatic-api/internal/repository/redis/session_repository"
	"postmatic-api/pkg/errs"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

// allowedFromBase: key -> base URL.
// from boleh berupa:
// - key: "dashboard"
// - key + path/query: "dashboard/xxx?foo=bar"
// - full url: "https://dashboard.postmatic.id/xxx"
var allowedFromBase = map[string]string{
	"postmatic.id":      "https://postmatic.id",
	"dashboard":         "https://dashboard.postmatic.id",
	"creator":           "https://creator.postmatic.id",
	"auth":              "https://auth.postmatic.id",
	"docs":              "https://docs.postmatic.id",
	"dashboard-staging": "https://dashboard-staging.postmatic.id",
	"landing-staging":   "https://landing-staging.postmatic.id",
	"creator-staging":   "https://creator-staging.postmatic.id",
	"auth-staging":      "https://auth-staging.postmatic.id",
	"docs-staging":      "https://docs-staging.postmatic.id",
}

type GoogleOAuthService struct {
	store            entity.Store
	queue            queue.MailerProducer
	cfg              config.Config
	sessionRepo      *sessRepo.SessionRepository
	emailLimiterRepo *emailLimiterRepo.LimiterEmailRepo
	conf             *oauth2.Config
	tm               token.TokenMaker
}

func NewService(
	store entity.Store,
	queue queue.MailerProducer,
	cfg config.Config,
	sessionRepo *sessRepo.SessionRepository,
	emailLimiterRepo *emailLimiterRepo.LimiterEmailRepo,
	tm token.TokenMaker,
) *GoogleOAuthService {
	oauthConf := cfg.GoogleOAuthConfig()
	return &GoogleOAuthService{
		store:            store,
		queue:            queue,
		cfg:              cfg,
		sessionRepo:      sessionRepo,
		emailLimiterRepo: emailLimiterRepo,
		conf:             oauthConf,
		tm:               tm,
	}
}

/* -----------------------------
   AUTH URL (untuk button FE)
------------------------------ */

func (s *GoogleOAuthService) GetGoogleAuthURL(ctx context.Context, from string) (GoogleOAuthAuthURLResponse, error) {
	// normalize + validate from (anti open-redirect)
	normFrom, err := s.normalizeFrom(from)
	if err != nil {
		return GoogleOAuthAuthURLResponse{}, errs.NewBadRequest("GOOGLE_OAUTH_INVALID_FROM")
	}

	// Scope harus include openid,email,profile di config
	// State kita sign agar bisa diverifikasi di callback
	state, err := s.signState(oauthStatePayload{
		From: normFrom,
		Exp:  time.Now().Add(10 * time.Minute).Unix(),
		N:    uuid.NewString(),
	})
	if err != nil {
		return GoogleOAuthAuthURLResponse{}, errs.NewInternalServerError(err)
	}

	// Offline TIDAK wajib kalau kamu cuma buat login.
	// Kalau butuh refresh token Google API, tambahkan oauth2.AccessTypeOffline dan prompt=consent.
	authURL := s.conf.AuthCodeURL(
		state,
		// oauth2.AccessTypeOffline,
		// oauth2.SetAuthURLParam("prompt", "consent"),
	)

	return GoogleOAuthAuthURLResponse{AuthURL: authURL}, nil
}

/* -----------------------------
   CALLBACK: code -> token -> id_token -> profile/user -> app tokens
------------------------------ */

func (s *GoogleOAuthService) LoginGoogleCallback(
	ctx context.Context,
	input GoogleOAuthCallbackInput,
	session SessionInput,
) (LoginGoogleResponse, error) {
	// 1) verify state
	st, err := s.verifyState(input.State)
	if err != nil {
		return LoginGoogleResponse{}, errs.NewBadRequest("GOOGLE_TOKEN_VERIFY_STATE_FAILED")
	}

	// ✅ from sumber kebenaran dari state (Google callback GET tidak membawa "from")
	// defense-in-depth: validate lagi supaya state tidak bisa mengarah ke domain lain
	normFrom, err := s.normalizeFrom(st.From)
	if err != nil {
		return LoginGoogleResponse{}, errs.NewBadRequest("GOOGLE_OAUTH_INVALID_FROM")
	}
	input.From = normFrom

	// 2) exchange code -> token
	tok, err := s.conf.Exchange(ctx, input.Code)
	if err != nil {
		return LoginGoogleResponse{}, errs.NewBadRequest("GOOGLE_TOKEN_EXCHANGE_FAILED")
	}

	rawIDToken, _ := tok.Extra("id_token").(string)
	if rawIDToken == "" {
		return LoginGoogleResponse{}, errs.NewBadRequest("GOOGLE_TOKEN_MISSING_ID_TOKEN")
	}

	// 3) validate id_token (ambil email, name, picture dari sini)
	payload, err := idtoken.Validate(ctx, rawIDToken, s.conf.ClientID)
	if err != nil {
		return LoginGoogleResponse{}, errs.NewBadRequest("GOOGLE_TOKEN_VALIDATION_FAILED")
	}

	email := getStringClaim(payload.Claims, "email")
	name := getStringClaim(payload.Claims, "name")
	picture := getStringClaim(payload.Claims, "picture")

	if email == "" || name == "" {
		return LoginGoogleResponse{}, errs.NewBadRequest("GOOGLE_TOKEN_MISSING_REQUIRED_CLAIMS")
	}

	// 4) cari/buat profile berdasar email (karena email hanya di profile)
	profile, err := s.store.GetProfileByEmail(ctx, email)
	if err != nil && err != sql.ErrNoRows {
		return LoginGoogleResponse{}, errs.NewInternalServerError(err)
	}

	// 5) jika profile ada, cari user provider google
	var targetUser entity.User
	userGoogleFound := false

	if profile.ID != uuid.Nil {
		users, err := s.store.ListUsersByProfileId(ctx, profile.ID)
		if err != nil {
			return LoginGoogleResponse{}, errs.NewInternalServerError(err)
		}
		for _, u := range users {
			if u.Provider == entity.AuthProviderGoogle {
				targetUser = u
				userGoogleFound = true
				break
			}
		}
	}

	// 6) create profile+user jika profile belum ada
	if profile.ID == uuid.Nil {
		e := s.store.ExecTx(ctx, func(q *entity.Queries) error {
			// (optional) simpan picture kalau ada
			img := sql.NullString{}
			if picture != "" {
				img = sql.NullString{Valid: true, String: picture}
			}

			prof, err := q.CreateProfile(ctx, entity.CreateProfileParams{
				Name:        name,
				Email:       email,
				ImageUrl:    img,
				CountryCode: sql.NullString{},
				Phone:       sql.NullString{},
				Description: sql.NullString{},
			})
			if err != nil {
				return err
			}
			profile = prof

			user, err := q.CreateUser(ctx, entity.CreateUserParams{
				ProfileID: prof.ID,
				Provider:  entity.AuthProviderGoogle,
				Password:  sql.NullString{},
			})
			if err != nil {
				return err
			}

			// ✅ pakai q.VerifyUser agar tetap dalam Tx
			_, err = q.VerifyUser(ctx, user.ID)
			if err != nil {
				return err
			}

			targetUser = user
			userGoogleFound = true
			return nil
		})
		if e != nil {
			return LoginGoogleResponse{}, errs.NewInternalServerError(e)
		}
		// ENQUEUE WELCOME EMAIL
		ctxQ, cancelQ := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelQ()
		s.queue.EnqueueWelcomeEmail(ctxQ, mailer.WelcomeInputDTO{
			Email: profile.Email,
			Name:  profile.Name,
			From:  normFrom,
		})
	}

	// 7) profile ada tapi belum ada user google -> buat user google
	if profile.ID != uuid.Nil && !userGoogleFound {
		e := s.store.ExecTx(ctx, func(q *entity.Queries) error {
			user, err := q.CreateUser(ctx, entity.CreateUserParams{
				ProfileID: profile.ID,
				Provider:  entity.AuthProviderGoogle,
				Password:  sql.NullString{},
			})
			if err != nil {
				return err
			}

			_, err = q.VerifyUser(ctx, user.ID)
			if err != nil {
				return err
			}

			targetUser = user
			return nil
		})
		if e != nil {
			return LoginGoogleResponse{}, errs.NewInternalServerError(e)
		}
		// ENQUEUE WELCOME EMAIL
		ctxQ, cancelQ := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelQ()
		s.queue.EnqueueWelcomeEmail(ctxQ, mailer.WelcomeInputDTO{
			Email: profile.Email,
			Name:  profile.Name,
			From:  normFrom,
		})
	}

	// 8) final imageUrl (kalau profile kosong tapi google ada picture, boleh update kemudian kalau kamu mau)
	var imageUrl *string
	if profile.ImageUrl.Valid {
		imageUrl = &profile.ImageUrl.String
	}

	// 9) issue token app kamu
	accessToken, err := s.tm.GenerateAccessToken(
		token.GenerateAccessTokenInput{
			ID:       targetUser.ID.String(),
			Email:    profile.Email,
			Name:     profile.Name,
			ImageUrl: imageUrl,
		},
	)
	if err != nil {
		return LoginGoogleResponse{}, errs.NewInternalServerError(err)
	}
	refreshToken, err := s.tm.GenerateRefreshToken(
		token.GenerateRefreshTokenInput{
			ID:    targetUser.ID.String(),
			Email: profile.Email,
		},
	)
	if err != nil {
		return LoginGoogleResponse{}, errs.NewInternalServerError(err)
	}

	// 10) save session
	sessionID := uuid.New().String()
	newSession := sessRepo.RedisSession{
		ID:           sessionID,
		RefreshToken: refreshToken,
		Browser:      session.DeviceInfo.Browser,
		Platform:     session.DeviceInfo.Platform,
		Device:       session.DeviceInfo.Device,
		ClientIP:     session.DeviceInfo.ClientIP,
		ProfileID:    profile.ID.String(),
		CreatedAt:    time.Now(),
		ExpiredAt:    time.Now().Add(s.cfg.JWT_REFRESH_TOKEN_EXPIRED),
	}
	s.sessionRepo.SaveSession(ctx, newSession, s.cfg.JWT_REFRESH_TOKEN_EXPIRED)

	return LoginGoogleResponse{
		ID:           profile.ID.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Name:         profile.Name,
		Email:        profile.Email,
		ImageUrl:     imageUrl,
		From:         input.From,
	}, nil
}

/* -----------------------------
   FROM NORMALIZER (allowlist)
------------------------------ */

// normalizeFrom menerima:
// - key ("dashboard")
// - key+path/query ("dashboard/xxx?foo=bar")
// - full URL ("https://dashboard.postmatic.id/xxx")
func (s *GoogleOAuthService) normalizeFrom(from string) (string, error) {
	from = strings.TrimSpace(from)
	if from == "" {
		return "", errors.New("empty from")
	}

	// Full URL mode
	if strings.HasPrefix(from, "http://") || strings.HasPrefix(from, "https://") {
		u, err := url.Parse(from)
		if err != nil {
			return "", err
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return "", errors.New("invalid scheme")
		}
		if u.Host == "" {
			return "", errors.New("missing host")
		}
		if u.User != nil {
			return "", errors.New("userinfo not allowed")
		}
		u.Fragment = "" // drop fragment

		if !s.isAllowedHost(u.Host) {
			return "", errors.New("host not allowed")
		}
		return u.String(), nil
	}

	// Key mode: <key>[/path][?query]
	key := from
	suffix := ""

	if i := strings.IndexAny(from, "/?"); i != -1 {
		key = from[:i]
		suffix = from[i:] // starts with "/" or "?"
	}

	base, ok := allowedFromBase[key]
	if !ok {
		return "", errors.New("unknown from key")
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	// Optional extra safety: even base mapping must be allowed
	if !s.isAllowedHost(baseURL.Host) {
		return "", errors.New("base host not allowed")
	}

	if suffix == "" {
		return baseURL.String(), nil
	}

	// Disallow protocol-relative path ("//evil.com")
	if strings.HasPrefix(suffix, "//") {
		return "", errors.New("invalid suffix")
	}

	su, err := url.Parse(suffix)
	if err != nil {
		return "", err
	}

	// Merge
	if su.Path != "" {
		// ensure leading slash
		if !strings.HasPrefix(su.Path, "/") {
			su.Path = "/" + su.Path
		}
		baseURL.Path = su.Path
	}
	baseURL.RawQuery = su.RawQuery
	baseURL.Fragment = "" // drop fragment

	return baseURL.String(), nil
}

func (s *GoogleOAuthService) isAllowedHost(host string) bool {
	// allow hosts from mapping
	for _, base := range allowedFromBase {
		u, err := url.Parse(base)
		if err != nil {
			continue
		}
		if strings.EqualFold(host, u.Host) {
			return true
		}
	}

	// optional: allow localhost for development
	if strings.HasPrefix(strings.ToLower(host), "localhost") || strings.HasPrefix(strings.ToLower(host), "127.0.0.1") {
		return true
	}

	return false
}

/* -----------------------------
   STATE SIGN/VERIFY (HMAC)
------------------------------ */

type oauthStatePayload struct {
	From string `json:"from"`
	Exp  int64  `json:"exp"`
	N    string `json:"n"` // nonce
}

func (s *GoogleOAuthService) oauthStateSecret() []byte {
	return []byte(s.cfg.OAUTH_GOOGLE_SECRET)
}

func (s *GoogleOAuthService) signState(p oauthStatePayload) (string, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	msg := base64.RawURLEncoding.EncodeToString(b)

	mac := hmac.New(sha256.New, s.oauthStateSecret())
	mac.Write([]byte(msg))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return msg + "." + sig, nil
}

func (s *GoogleOAuthService) verifyState(state string) (oauthStatePayload, error) {
	parts := strings.Split(state, ".")
	if len(parts) != 2 {
		return oauthStatePayload{}, errors.New("bad format")
	}
	msg, sig := parts[0], parts[1]

	mac := hmac.New(sha256.New, s.oauthStateSecret())
	mac.Write([]byte(msg))
	expect := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expect)) {
		return oauthStatePayload{}, errors.New("bad signature")
	}

	raw, err := base64.RawURLEncoding.DecodeString(msg)
	if err != nil {
		return oauthStatePayload{}, err
	}

	var p oauthStatePayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return oauthStatePayload{}, err
	}
	if p.Exp <= time.Now().Unix() {
		return oauthStatePayload{}, errors.New("expired")
	}
	if p.N == "" || p.From == "" {
		return oauthStatePayload{}, errors.New("missing fields")
	}
	return p, nil
}

func getStringClaim(claims map[string]any, key string) string {
	v, ok := claims[key]
	if !ok || v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}
