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
	"strings"
	"time"

	"postmatic-api/config"
	"postmatic-api/internal/module/headless/mailer"
	"postmatic-api/internal/repository/entity"
	emailLimiterRepo "postmatic-api/internal/repository/redis/email_limiter_repository"
	sessRepo "postmatic-api/internal/repository/redis/session_repository"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/token"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

type GoogleOAuthService struct {
	store            entity.Store
	mailer           mailer.MailerService
	cfg              config.Config
	sessionRepo      *sessRepo.SessionRepository
	emailLimiterRepo *emailLimiterRepo.LimiterEmailRepo
	conf             *oauth2.Config
}

func NewService(
	store entity.Store,
	mailer mailer.MailerService,
	cfg config.Config,
	sessionRepo *sessRepo.SessionRepository,
	emailLimiterRepo *emailLimiterRepo.LimiterEmailRepo,
) *GoogleOAuthService {
	oauthConf := cfg.GoogleOAuthConfig()
	return &GoogleOAuthService{
		store:            store,
		mailer:           mailer,
		cfg:              cfg,
		sessionRepo:      sessionRepo,
		emailLimiterRepo: emailLimiterRepo,
		conf:             oauthConf,
	}
}

/* -----------------------------
   AUTH URL (untuk button FE)
------------------------------ */

func (s *GoogleOAuthService) GetGoogleAuthURL(ctx context.Context, from string) (GoogleOAuthAuthURLResponse, error) {
	// Scope harus include openid,email,profile di config
	// State kita sign agar bisa diverifikasi di callback
	state, err := s.signState(oauthStatePayload{
		From: from,
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
	input.From = st.From
	// Optional: pastikan "from" sama dengan yang ada di state
	if st.From != input.From {
		return LoginGoogleResponse{}, errs.NewBadRequest("GOOGLE_TOKEN_STATE_MISMATCH")
	}

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
		// TODO: send welcome email
		if e != nil {
			return LoginGoogleResponse{}, errs.NewInternalServerError(e)
		}
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
		// TODO: send welcome email
		if e != nil {
			return LoginGoogleResponse{}, errs.NewInternalServerError(e)
		}
	}

	// 8) final imageUrl (kalau profile kosong tapi google ada picture, boleh update kemudian kalau kamu mau)
	var imageUrl *string
	if profile.ImageUrl.Valid {
		imageUrl = &profile.ImageUrl.String
	}

	// 9) issue token app kamu
	accessToken, err := token.GenerateAccessToken(targetUser.ID.String(), profile.Email, profile.Name, imageUrl)
	if err != nil {
		return LoginGoogleResponse{}, errs.NewInternalServerError(err)
	}
	refreshToken, err := token.GenerateRefreshToken(targetUser.ID.String(), profile.Email)
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
