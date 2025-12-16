// internal/module/account/auth/service.go
package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	"postmatic-api/config"
	"postmatic-api/internal/module/headless/mailer"
	"postmatic-api/internal/repository/entity"
	repositoryRedis "postmatic-api/internal/repository/redis"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/token"
	"postmatic-api/pkg/utils"

	"github.com/google/uuid"
)

type AuthService struct {
	store       entity.Store
	mailer      mailer.MailerService
	cfg         config.Config
	sessionRepo *repositoryRedis.SessionRepository
}

// Update Constructor: Minta Token Maker dari main.go
func NewService(store entity.Store, mailer mailer.MailerService, cfg config.Config, sessionRepo *repositoryRedis.SessionRepository) *AuthService {
	return &AuthService{
		store:       store,
		mailer:      mailer,
		cfg:         cfg,
		sessionRepo: sessionRepo,
	}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (RegisterResponse, error) {

	hashedPassword, e := utils.HashPassword(input.Password)
	if e != nil {
		return RegisterResponse{}, errs.NewInternalServerError(e)
	}

	var finalProfile entity.Profile // Variable untuk menampung hasil dari dalam transaksi

	// PANGGIL TRANSAKSI
	// Semua yang ada di dalam fungsi ini bersifat ATOMIC (All or Nothing)
	err := s.store.ExecTx(ctx, func(q *entity.Queries) error {

		// A. Cek User (Gunakan 'q', bukan 's.store')
		checkUser, err := q.GetUserByEmailProfile(ctx, input.Email)
		for _, u := range checkUser {
			if u.Provider == "credentials" {
				return errs.NewBadRequest("EMAIL_ALREADY_EXISTS")
			}
		}

		// B. Cek Profile
		profile, err := q.GetProfileByEmail(ctx, input.Email)

		// Logic Profile Baru vs Lama
		if err == sql.ErrNoRows {
			newProfile, err := q.CreateProfile(ctx, entity.CreateProfileParams{
				Name:        input.Name,
				Email:       input.Email,
				ImageUrl:    sql.NullString{},
				CountryCode: sql.NullString{},
				Phone:       sql.NullString{},
				Description: sql.NullString{},
			})
			if err != nil {
				return err // Otomatis Rollback
			}
			profile = newProfile
		} else if err != nil {
			return err // Otomatis Rollback
		}

		// C. Buat User (Sekarang aman, profile pasti ada)
		_, err = q.CreateUser(ctx, entity.CreateUserParams{
			ProfileID: profile.ID,
			Password:  sql.NullString{String: hashedPassword, Valid: true},
			Provider:  "credentials",
		})

		if err != nil {
			return err // Otomatis Rollback (Profile yang tadi dibuat juga akan hilang)
		}

		createAccountToken, err := token.GenerateCreateAccountToken(profile.ID.String(), profile.Email, profile.Name, nil)
		if err != nil {
			return err // Otomatis Rollback (Profile yang tadi dibuat juga akan hilang)
		}

		err = s.sendVerificationEmail(ctx, profile.Name, profile.Email, createAccountToken, input.From)
		if err != nil {
			return err // Otomatis Rollback (Profile yang tadi dibuat juga akan hilang)
		}

		// Assign ke variable luar agar bisa di-return
		finalProfile = profile
		return nil // Commit Transaksi
	})

	if err != nil {
		// Error sudah dibungkus dari dalam ExecTx
		return RegisterResponse{}, err
	}

	// TODO: add bucket redis for retry
	retryAfter := s.cfg.CAN_RESEND_EMAIL_AFTER

	return RegisterResponse{
		ID:         finalProfile.ID.String(),
		Name:       finalProfile.Name,
		Email:      finalProfile.Email,
		ImageUrl:   nil,
		RetryAfter: retryAfter,
	}, nil
}

func (s *AuthService) LoginCredentials(ctx context.Context, input LoginCredentialsInput, session SessionInput) (LoginResponse, error) {

	// 1. Ambil Profile (Read Only - Tidak perlu Tx)
	profile, err := s.store.GetProfileByEmail(ctx, input.Email)
	if err == sql.ErrNoRows {
		return LoginResponse{}, errs.NewUnauthorized("EMAIL_NOT_FOUND")
	}
	if err != nil {
		return LoginResponse{}, errs.NewInternalServerError(err)
	}

	// 2. Ambil User List
	users, err := s.store.ListUsersByProfileId(ctx, profile.ID)
	if err != nil {
		return LoginResponse{}, errs.NewInternalServerError(err)
	}

	var targetUser entity.User
	userCredFound := false

	// Cari user dengan provider 'credentials'
	for _, u := range users {
		if u.Provider == "credentials" {
			targetUser = u
			userCredFound = true
			break
		}
	}

	// LOGIC: Jika user credentials belum ada, tapi input password -> Create User (Link Account)
	if !userCredFound {
		hashedPassword, err := utils.HashPassword(input.Password)
		if err != nil {
			return LoginResponse{}, errs.NewInternalServerError(err)
		}

		// Single Write - Cukup panggil store langsung (Auto Commit)
		createdUser, err := s.store.CreateUser(ctx, entity.CreateUserParams{
			ProfileID: profile.ID,
			Password:  sql.NullString{String: hashedPassword, Valid: true},
			Provider:  "credentials",
		})
		if err != nil {
			return LoginResponse{}, errs.NewInternalServerError(err)
		}
		targetUser = createdUser
	} else {
		// Jika user ada, COMPARE PASSWORD
		if !targetUser.Password.Valid || !utils.ComparePassword(targetUser.Password.String, input.Password) {
			return LoginResponse{}, errs.NewUnauthorized("INVALID_CREDENTIALS")
		}
	}

	if !targetUser.VerifiedAt.Valid {
		createAccountToken, err := token.GenerateCreateAccountToken(profile.ID.String(), profile.Email, profile.Name, nil)
		if err != nil {
			return LoginResponse{}, errs.NewInternalServerError(err)
		}
		err = s.sendVerificationEmail(ctx, profile.Name, profile.Email, createAccountToken, input.From)
		if err != nil {
			return LoginResponse{}, errs.NewInternalServerError(err)
		}
		return LoginResponse{}, errs.NewUnauthorized("EMAIL_NOT_VERIFIED")
	}

	// 3. Generate Tokens
	// Convert UUID ke String
	pID := profile.ID.String()

	var imageUrl *string
	if profile.ImageUrl.Valid {
		imageUrl = &profile.ImageUrl.String
	}

	accessToken, err := token.GenerateAccessToken(pID, profile.Email, profile.Name, imageUrl)
	if err != nil {
		return LoginResponse{}, errs.NewInternalServerError(err)
	}

	refreshToken, err := token.GenerateRefreshToken(pID, profile.Email, profile.Name, imageUrl)
	if err != nil {
		return LoginResponse{}, errs.NewInternalServerError(err)
	}

	sessionID := uuid.New().String()

	newSession := repositoryRedis.RedisSession{
		ID:           sessionID,
		RefreshToken: refreshToken,
		Browser:      session.DeviceInfo.Browser,
		Platform:     session.DeviceInfo.Platform, // OS
		Device:       session.DeviceInfo.Device,
		ClientIP:     session.DeviceInfo.ClientIP,
		ProfileID:    pID,
		CreatedAt:    time.Now(),
		ExpiredAt:    time.Now().Add(s.cfg.JWT_REFRESH_TOKEN_EXPIRED),
	}

	err = s.sessionRepo.SaveSession(ctx, newSession, s.cfg.JWT_REFRESH_TOKEN_EXPIRED)
	if err != nil {
		// Jika gagal simpan session, login harus dianggap gagal
		return LoginResponse{}, errs.NewInternalServerError(err)
	}

	// 4. Return Response Lengkap
	return LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ID:           pID,
		Name:         profile.Name,
		Email:        profile.Email,
		ImageUrl:     imageUrl,
	}, nil
}
func (s *AuthService) RefreshToken(ctx context.Context, input RefreshTokenInput) (LoginResponse, error) {
	// 1. Validasi Signature JWT
	valid, err := token.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		return LoginResponse{}, errs.NewUnauthorized("INVALID_REFRESH_TOKEN")
	}

	// 2. Cek Keberadaan Session di Redis (Whitelist Check)
	// Ingat: GetSessionByRefreshToken Anda sebelumnya me-return POINTER (*RedisSession)
	sess, err := s.sessionRepo.GetSessionByRefreshToken(ctx, valid.ID, input.RefreshToken)
	if err != nil {
		return LoginResponse{}, errs.NewInternalServerError(err)
	}
	if sess == nil {
		// Token valid secara signature, tapi tidak ada di Redis (sudah logout/revoked)
		return LoginResponse{}, errs.NewUnauthorized("SESSION_EXPIRED_OR_REVOKED")
	}

	// 3. Generate Access Token Baru (Selalu dilakukan)
	accessToken, err := token.GenerateAccessToken(valid.ID, valid.Email, valid.Name, valid.ImageUrl)
	if err != nil {
		return LoginResponse{}, errs.NewInternalServerError(err)
	}

	// 4. Logic Rotasi Refresh Token (Jika perlu)
	currentRefreshToken := input.RefreshToken

	// Hitung sisa waktu hidup token
	timeLeft := time.Until(sess.ExpiredAt)

	// Jika sisa waktu kurang dari batas renewal (misal kurang dari 1 hari), ROTASI TOKEN
	if timeLeft < s.cfg.JWT_REFRESH_TOKEN_RENEWAL {
		// A. Generate Token Baru
		newRefreshToken, err := token.GenerateRefreshToken(valid.ID, valid.Email, valid.Name, valid.ImageUrl)
		if err != nil {
			return LoginResponse{}, errs.NewInternalServerError(err)
		}

		// B. Update Object Session (Memory)
		sess.RefreshToken = newRefreshToken

		// C. Perpanjang ExpiredAt (Memory)
		// Reset umur session menjadi full lagi (misal 7 hari lagi dari sekarang)
		sess.ExpiredAt = time.Now().Add(s.cfg.JWT_REFRESH_TOKEN_EXPIRED)

		// D. SIMPAN KE REDIS (PENTING!!!)
		// Kita overwrite session ID yang sama dengan data baru & TTL baru
		err = s.sessionRepo.SaveSession(ctx, *sess, s.cfg.JWT_REFRESH_TOKEN_EXPIRED)
		if err != nil {
			return LoginResponse{}, errs.NewInternalServerError(err)
		}

		// E. Update variable return
		currentRefreshToken = newRefreshToken
	}

	return LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: currentRefreshToken, // Bisa token lama, bisa token baru
		ID:           valid.ID,
		Name:         valid.Name,
		Email:        valid.Email,
		ImageUrl:     valid.ImageUrl,
	}, nil
}

// FOR GET THERE'S NO STORE IN DB (ONLY CHECK FOR UI)
func (s *AuthService) CheckVerifyToken(ctx context.Context, input string) (VerifyCreateAccountTokenResponse, error) {
	valid, err := token.ValidateCreateAccountToken(input)
	if err != nil {
		// Token Invalid
		return VerifyCreateAccountTokenResponse{
			ID:    nil,
			Name:  nil,
			Email: nil,
			Valid: false,
		}, errs.NewBadRequest("INVALID_CREATE_ACCOUNT_TOKEN")
	}
	decoded, err := token.DecodeTokenWithoutVerify(input)
	if err != nil || decoded == nil {
		// Token Invalid
		return VerifyCreateAccountTokenResponse{
			ID:    nil,
			Name:  nil,
			Email: nil,
			Valid: false,
		}, errs.NewBadRequest("INVALID_CREATE_ACCOUNT_TOKEN")
	}

	if decoded.ID == "" {
		// Token Invalid
		return VerifyCreateAccountTokenResponse{
			ID:    nil,
			Name:  nil,
			Email: nil,
			Valid: false,
		}, errs.NewBadRequest("INVALID_CREATE_ACCOUNT_TOKEN")
	}

	profileId, err := uuid.Parse(decoded.ID)

	if err != nil || profileId == uuid.Nil {
		// Token Invalid
		return VerifyCreateAccountTokenResponse{
			ID:    nil,
			Name:  nil,
			Email: nil,
			Valid: false,
		}, errs.NewBadRequest("INVALID_CREATE_ACCOUNT_TOKEN")
	}

	users, err := s.store.ListUsersByProfileId(ctx, profileId)
	if err != nil {
		return VerifyCreateAccountTokenResponse{
			ID:    nil,
			Name:  nil,
			Email: nil,
			Valid: false,
		}, errs.NewInternalServerError(err)
	}

	credUser := entity.User{}
	for _, u := range users {
		if u.Provider == "credentials" {
			credUser = u
			break
		}
	}

	if credUser.ID == uuid.Nil {
		return VerifyCreateAccountTokenResponse{
			ID:    nil,
			Name:  nil,
			Email: nil,
			Valid: false,
		}, errs.NewBadRequest("USER_NOT_FOUND")
	}

	if credUser.VerifiedAt.Valid {
		return VerifyCreateAccountTokenResponse{
			ID:    nil,
			Name:  nil,
			Email: nil,
			Valid: false,
		}, errs.NewBadRequest("USER_ALREADY_VERIFIED")
	}

	return VerifyCreateAccountTokenResponse{
		ID:       &valid.ID,
		Name:     &valid.Name,
		Email:    &valid.Email,
		ImageUrl: valid.ImageUrl,
		Valid:    true,
	}, nil
}

func (s *AuthService) SubmitVerifyToken(ctx context.Context, input SubmitVerifyTokenInput, session SessionInput) (VerifyCreateAccountResponse, error) {
	valid, err := s.CheckVerifyToken(ctx, input.Token)
	if err != nil {
		return VerifyCreateAccountResponse{}, err
	}

	if valid.ID == nil {
		return VerifyCreateAccountResponse{}, errs.NewBadRequest("INVALID_CREATE_ACCOUNT_TOKEN")
	}

	profileId, err := uuid.Parse(*valid.ID)
	if err != nil {
		return VerifyCreateAccountResponse{}, errs.NewBadRequest("INVALID_PROFILE_ID")
	}

	users, err := s.store.ListUsersByProfileId(ctx, profileId)
	if err != nil {
		return VerifyCreateAccountResponse{}, errs.NewInternalServerError(err)
	}

	var user entity.User

	for _, u := range users {
		if u.Provider == "credentials" {
			user = u
			break
		}
	}

	if user.ID == uuid.Nil {
		return VerifyCreateAccountResponse{}, errs.NewInternalServerError(errors.New("USER_NOT_FOUND"))
	}

	if user.VerifiedAt.Valid {
		return VerifyCreateAccountResponse{}, errs.NewBadRequest("USER_ALREADY_VERIFIED")
	}

	// Single Write - Cukup panggil store langsung (Auto Commit)
	_, err = s.store.VerifyUser(ctx, user.ID)
	if err != nil {
		return VerifyCreateAccountResponse{}, errs.NewInternalServerError(err)
	}

	var imageUrl *string
	if valid.ImageUrl != nil && *valid.ImageUrl != "" {
		imageUrl = valid.ImageUrl
	}

	accessToken, err := token.GenerateAccessToken(profileId.String(), *valid.Email, *valid.Name, imageUrl)
	if err != nil {
		return VerifyCreateAccountResponse{}, errs.NewInternalServerError(err)
	}

	refreshToken, err := token.GenerateRefreshToken(profileId.String(), *valid.Email, *valid.Name, imageUrl)
	if err != nil {
		return VerifyCreateAccountResponse{}, errs.NewInternalServerError(err)
	}

	// GO ROUTINE FOR SENDING EMAIL IN BACKGROUND
	go func() {
		bgCtx := context.Background()

		link := s.cfg.AUTH_URL + "&from=" + url.QueryEscape(input.From)

		sessionID := uuid.New().String()

		newSession := repositoryRedis.RedisSession{
			ID:           sessionID,
			RefreshToken: refreshToken,
			Browser:      session.DeviceInfo.Browser,
			Platform:     session.DeviceInfo.Platform, // OS
			Device:       session.DeviceInfo.Device,
			ClientIP:     session.DeviceInfo.ClientIP,
			ProfileID:    profileId.String(),
			CreatedAt:    time.Now(),
			ExpiredAt:    time.Now().Add(s.cfg.JWT_REFRESH_TOKEN_EXPIRED),
		}

		err = s.sessionRepo.SaveSession(ctx, newSession, s.cfg.JWT_REFRESH_TOKEN_EXPIRED)

		// Data yang mau dikirim ke HTML {{.Name}}, {{.Email}}, {{.Link}}
		templateData := mailer.WelcomeInput{
			Name:  *valid.Name,
			Email: *valid.Email,
			Link:  link,
		}

		emailInput := mailer.SendEmailInput{
			To:           *valid.Email,
			Subject:      "Selamat Datang di Postmatic!",
			TemplateName: "welcome.html",
			Data:         templateData,
		}

		if err := s.mailer.SendEmail(bgCtx, emailInput); err != nil {
			println("Failed to send email:", err.Error())
		}
	}()

	return VerifyCreateAccountResponse{
		ID:           valid.ID,
		Name:         valid.Name,
		Email:        valid.Email,
		ImageUrl:     valid.ImageUrl,
		Valid:        true,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) ResendEmailVerification(ctx context.Context, input ResendEmailVerificationInput) (ResendEmailVerificationResponse, error) {
	email := input.Email

	users, err := s.store.GetUserByEmailProfile(ctx, email)
	if err != nil {
		return ResendEmailVerificationResponse{}, errs.NewInternalServerError(err)
	}

	if len(users) == 0 {
		return ResendEmailVerificationResponse{}, errs.NewBadRequest("USER_NOT_FOUND")
	}

	var user entity.User
	var profile entity.Profile
	for _, u := range users {
		if u.Provider == "credentials" {
			user = entity.User{
				ID:         u.ID,
				VerifiedAt: u.VerifiedAt,
			}
			profile = entity.Profile{
				ID:       u.ProfileID,
				Name:     u.Name,
				Email:    u.Email,
				ImageUrl: u.ImageUrl,
			}
			break
		}
	}

	if user.ID == uuid.Nil {
		return ResendEmailVerificationResponse{}, errs.NewBadRequest("USER_NOT_FOUND")
	}

	if user.VerifiedAt.Valid {
		return ResendEmailVerificationResponse{}, errs.NewBadRequest("USER_ALREADY_VERIFIED")
	}

	if profile.ID == uuid.Nil {
		return ResendEmailVerificationResponse{}, errs.NewBadRequest("USER_NOT_FOUND")
	}

	var imageUrl *string
	if profile.ImageUrl.Valid {
		imageUrl = &profile.ImageUrl.String
	}

	token, err := token.GenerateCreateAccountToken(user.ID.String(), profile.Email, profile.Name, imageUrl)
	if err != nil {
		return ResendEmailVerificationResponse{}, errs.NewInternalServerError(err)
	}

	err = s.sendVerificationEmail(ctx, profile.Name, profile.Email, token, input.From)
	if err != nil {
		return ResendEmailVerificationResponse{}, err
	}

	// TODO: add bucket redis for retry
	retryAfter := s.cfg.CAN_RESEND_EMAIL_AFTER

	return ResendEmailVerificationResponse{
		ID:         profile.ID.String(),
		Name:       profile.Name,
		Email:      profile.Email,
		ImageUrl:   imageUrl,
		RetryAfter: retryAfter,
	}, nil

}

func (s *AuthService) sendVerificationEmail(ctx context.Context, name, to, token, from string) error {
	u, err := url.Parse(s.cfg.AUTH_URL + s.cfg.VERIFY_EMAIL_ROUTE + "/" + token)
	if err != nil {
		return errs.NewInternalServerError(err)
	}

	q := u.Query()
	q.Set("from", from)
	u.RawQuery = q.Encode()

	confirmUrl := u.String()
	fmt.Println(confirmUrl)
	templateData := mailer.VerificationInput{
		Name:       name,
		ConfirmUrl: confirmUrl,
	}

	err = s.mailer.SendEmail(ctx, mailer.SendEmailInput{
		To:           to,
		Subject:      "Konfirmasi Pendaftaran Akun",
		TemplateName: "verification.html",
		Type:         "html",
		Data:         templateData,
	})
	if err != nil {
		return errs.NewInternalServerError(err)
	}
	return nil
}
