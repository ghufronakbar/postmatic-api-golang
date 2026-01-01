// internal/module/account/auth/service.go
package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"postmatic-api/config"
	"postmatic-api/internal/module/headless/mailer"
	"postmatic-api/internal/module/headless/queue"
	"postmatic-api/internal/module/headless/token"
	"postmatic-api/internal/repository/entity"
	emailLimiterRepo "postmatic-api/internal/repository/redis/email_limiter_repository"
	sessRepo "postmatic-api/internal/repository/redis/session_repository"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/logger"
	"postmatic-api/pkg/utils"

	"github.com/google/uuid"
)

type AuthService struct {
	store            entity.Store
	queue            queue.MailerProducer
	cfg              config.Config
	sessionRepo      *sessRepo.SessionRepository
	emailLimiterRepo *emailLimiterRepo.LimiterEmailRepo
	tm               token.TokenMaker
}

// Update Constructor: Minta Token Maker dari main.go
func NewService(store entity.Store, queue queue.MailerProducer, cfg config.Config, sessionRepo *sessRepo.SessionRepository, emailLimiterRepo *emailLimiterRepo.LimiterEmailRepo, tm token.TokenMaker) *AuthService {
	return &AuthService{
		store:            store,
		queue:            queue,
		cfg:              cfg,
		sessionRepo:      sessionRepo,
		emailLimiterRepo: emailLimiterRepo,
		tm:               tm,
	}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (RegisterResponse, error) {

	hashedPassword, e := utils.HashPassword(input.Password)
	if e != nil {
		return RegisterResponse{}, errs.NewInternalServerError(e)
	}

	var finalProfile entity.Profile // Variable untuk menampung hasil dari dalam transaksi
	var createAccountToken string

	// PANGGIL TRANSAKSI
	// Semua yang ada di dalam fungsi ini bersifat ATOMIC (All or Nothing)
	err := s.store.ExecTx(ctx, func(q *entity.Queries) error {

		// A. Cek User (Gunakan 'q', bukan 's.store')
		checkUser, e := q.GetUserByEmailProfile(ctx, input.Email)
		if e != nil {
			return e
		}
		for _, u := range checkUser {
			if u.Provider == entity.AuthProviderCredential {
				return errs.NewBadRequest("EMAIL_ALREADY_EXISTS")
			}
		}

		// B. Cek Profile
		profile, e := q.GetProfileByEmail(ctx, input.Email)

		// Logic Profile Baru vs Lama
		if e == sql.ErrNoRows {
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
		} else if e != nil {
			return e // Otomatis Rollback
		}

		// C. Buat User (Sekarang aman, profile pasti ada)
		_, e = q.CreateUser(ctx, entity.CreateUserParams{
			ProfileID: profile.ID,
			Password:  sql.NullString{String: hashedPassword, Valid: true},
			Provider:  entity.AuthProviderCredential,
		})

		if e != nil {
			return e // Otomatis Rollback (Profile yang tadi dibuat juga akan hilang)
		}

		createAccountToken, e = s.tm.GenerateCreateAccountToken(token.GenerateCreateAccountTokenInput{
			ID:       profile.ID,
			Email:    profile.Email,
			Name:     profile.Name,
			ImageUrl: nil,
		})
		if e != nil {
			return e // Otomatis Rollback (Profile yang tadi dibuat juga akan hilang)
		}

		// Assign ke variable luar agar bisa di-return
		finalProfile = profile
		return nil // Commit Transaksi
	})

	if err != nil {
		// Error sudah dibungkus dari dalam ExecTx
		return RegisterResponse{}, err
	}

	// SEND EMAIL IN QUEUE
	ctxQ, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	e = s.queue.EnqueueUserVerification(ctxQ, mailer.VerificationInputDTO{
		Name:  finalProfile.Name,
		To:    finalProfile.Email,
		Token: createAccountToken,
		From:  input.From,
	})
	if e != nil {
		logger.From(ctx).Warn("enqueue user verification failed", "err", e)
	}

	//  SET RATE LIMITER BARU
	// Gunakan duration dari Config
	retryAfter := s.cfg.CAN_RESEND_EMAIL_AFTER // misal 60 detik
	retryAfterDuration := time.Duration(retryAfter) * time.Second

	// Simpan ke Redis (Hanya Email & TTL)
	err = s.emailLimiterRepo.SaveLimiterEmail(ctx, input.Email, retryAfterDuration)
	if err != nil {
		logger.From(ctx).Warn("failed to save email limiter", "err", err)
	}

	return RegisterResponse{
		ID:         finalProfile.ID,
		Name:       finalProfile.Name,
		Email:      finalProfile.Email,
		ImageUrl:   nil,
		RetryAfter: retryAfter,
	}, nil
}

func (s *AuthService) LoginCredential(ctx context.Context, input LoginCredentialInput, session SessionInput) (LoginResponse, error) {

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

	// Cari user dengan provider 'credential'
	for _, u := range users {
		if u.Provider == entity.AuthProviderCredential {
			targetUser = u
			userCredFound = true
			break
		}
	}

	// LOGIC: Jika user credential belum ada, tapi input password -> Create User (Link Account)
	if !userCredFound {
		hashedPassword, err := utils.HashPassword(input.Password)
		if err != nil {
			return LoginResponse{}, errs.NewInternalServerError(err)
		}

		// Single Write - Cukup panggil store langsung (Auto Commit)
		createdUser, err := s.store.CreateUser(ctx, entity.CreateUserParams{
			ProfileID: profile.ID,
			Password:  sql.NullString{String: hashedPassword, Valid: true},
			Provider:  entity.AuthProviderCredential,
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
		// A. CEK LIMITER DULU
		checkLimiter, _ := s.emailLimiterRepo.GetLimiterEmail(ctx, profile.Email)

		if checkLimiter != nil {
			return LoginResponse{
				Email:      input.Email,
				RetryAfter: checkLimiter.RetryAfterSeconds,
				ID:         profile.ID,
				Name:       profile.Name,
				ImageUrl:   nil,
			}, errs.NewBadRequest("PLEASE_WAIT")
		}

		// Jika belum kena limit, baru kirim email
		createAccountToken, err := s.tm.GenerateCreateAccountToken(
			token.GenerateCreateAccountTokenInput{
				ID:       profile.ID,
				Email:    profile.Email,
				Name:     profile.Name,
				ImageUrl: nil,
			},
		)
		if err != nil {
			return LoginResponse{}, errs.NewInternalServerError(err)
		}

		// Enqueue User Verification
		ctxQ, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		err = s.queue.EnqueueUserVerification(ctxQ, mailer.VerificationInputDTO{
			Name:  profile.Name,
			To:    profile.Email,
			Token: createAccountToken,
			From:  input.From,
		})
		if err != nil {
			logger.From(ctx).Warn("enqueue user verification failed", "err", err)
		}

		// SET LIMITER (PENTING)
		retryAfter := s.cfg.CAN_RESEND_EMAIL_AFTER
		_ = s.emailLimiterRepo.SaveLimiterEmail(ctx, profile.Email, time.Duration(retryAfter)*time.Second)

		return LoginResponse{
			Email:      input.Email,
			RetryAfter: retryAfter,
			ID:         profile.ID,
			Name:       profile.Name,
			ImageUrl:   nil,
		}, errs.NewUnauthorized("EMAIL_NOT_VERIFIED")
	}

	// 3. Generate Tokens
	pID := profile.ID

	var imageUrl *string
	if profile.ImageUrl.Valid {
		imageUrl = &profile.ImageUrl.String
	}

	accessToken, err := s.tm.GenerateAccessToken(
		token.GenerateAccessTokenInput{
			ID:       pID,
			Email:    profile.Email,
			Name:     profile.Name,
			ImageUrl: imageUrl,
		},
	)
	if err != nil {
		return LoginResponse{}, errs.NewInternalServerError(err)
	}

	refreshToken, err := s.tm.GenerateRefreshToken(
		token.GenerateRefreshTokenInput{
			ID:    pID,
			Email: profile.Email,
		},
	)
	if err != nil {
		return LoginResponse{}, errs.NewInternalServerError(err)
	}

	sessionID := uuid.New()

	newSession := sessRepo.RedisSession{
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
		RetryAfter:   0,
	}, nil
}
func (s *AuthService) RefreshToken(ctx context.Context, input RefreshTokenInput) (LoginResponse, error) {
	// 1. Validasi Signature JWT
	valid, err := s.tm.ValidateRefreshToken(input.RefreshToken)
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

	profUUID := valid.ID

	profile, err := s.store.GetProfileById(ctx, profUUID)
	if err != nil {
		return LoginResponse{}, errs.NewInternalServerError(err)
	}

	var imageUrl *string
	if profile.ImageUrl.Valid {
		imageUrl = &profile.ImageUrl.String
	}

	// 3. Generate Access Token Baru (Selalu dilakukan)
	accessToken, err := s.tm.GenerateAccessToken(
		token.GenerateAccessTokenInput{
			ID:       valid.ID,
			Email:    valid.Email,
			Name:     profile.Name,
			ImageUrl: imageUrl,
		},
	)
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
		newRefreshToken, err := s.tm.GenerateRefreshToken(
			token.GenerateRefreshTokenInput{
				ID:    valid.ID,
				Email: valid.Email,
			},
		)
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
	valid, err := s.tm.ValidateCreateAccountToken(input)
	if err != nil {
		// Token Invalid
		return VerifyCreateAccountTokenResponse{
			ID:    nil,
			Name:  nil,
			Email: nil,
			Valid: false,
		}, errs.NewBadRequest("INVALID_CREATE_ACCOUNT_TOKEN")
	}
	decoded, err := s.tm.DecodeTokenWithoutVerify(input)
	if err != nil || decoded == nil {
		// Token Invalid
		return VerifyCreateAccountTokenResponse{
			ID:    nil,
			Name:  nil,
			Email: nil,
			Valid: false,
		}, errs.NewBadRequest("INVALID_CREATE_ACCOUNT_TOKEN")
	}

	if decoded.ID == uuid.Nil {
		// Token Invalid
		return VerifyCreateAccountTokenResponse{
			ID:    nil,
			Name:  nil,
			Email: nil,
			Valid: false,
		}, errs.NewBadRequest("INVALID_CREATE_ACCOUNT_TOKEN")
	}

	if decoded.ID == uuid.Nil {
		// Token Invalid
		return VerifyCreateAccountTokenResponse{
			ID:    nil,
			Name:  nil,
			Email: nil,
			Valid: false,
		}, errs.NewBadRequest("INVALID_CREATE_ACCOUNT_TOKEN")
	}

	users, err := s.store.ListUsersByProfileId(ctx, decoded.ID)
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
		if u.Provider == entity.AuthProviderCredential {
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

	profileId := *valid.ID

	users, err := s.store.ListUsersByProfileId(ctx, profileId)
	if err != nil {
		return VerifyCreateAccountResponse{}, errs.NewInternalServerError(err)
	}

	var user entity.User

	for _, u := range users {
		if u.Provider == entity.AuthProviderCredential {
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

	accessToken, err := s.tm.GenerateAccessToken(
		token.GenerateAccessTokenInput{
			ID:       profileId,
			Email:    *valid.Email,
			Name:     *valid.Name,
			ImageUrl: imageUrl,
		},
	)
	if err != nil {
		return VerifyCreateAccountResponse{}, errs.NewInternalServerError(err)
	}

	refreshToken, err := s.tm.GenerateRefreshToken(
		token.GenerateRefreshTokenInput{
			ID:    profileId,
			Email: *valid.Email,
		},
	)
	if err != nil {
		return VerifyCreateAccountResponse{}, errs.NewInternalServerError(err)
	}

	sessionID := uuid.New()

	newSession := sessRepo.RedisSession{
		ID:           sessionID,
		RefreshToken: refreshToken,
		Browser:      session.DeviceInfo.Browser,
		Platform:     session.DeviceInfo.Platform, // OS
		Device:       session.DeviceInfo.Device,
		ClientIP:     session.DeviceInfo.ClientIP,
		ProfileID:    profileId,
		CreatedAt:    time.Now(),
		ExpiredAt:    time.Now().Add(s.cfg.JWT_REFRESH_TOKEN_EXPIRED),
	}

	err = s.sessionRepo.SaveSession(ctx, newSession, s.cfg.JWT_REFRESH_TOKEN_EXPIRED)
	if err != nil {
		return VerifyCreateAccountResponse{}, errs.NewInternalServerError(err)
	}

	ctxQ, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := s.queue.EnqueueWelcomeEmail(ctxQ, mailer.WelcomeInputDTO{
		Name:  *valid.Name,
		Email: *valid.Email,
		From:  input.From,
	}); err != nil {
		logger.From(ctx).Warn("enqueue welcome email failed", "err", err)
	}

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

	// 1. CEK RATE LIMITER
	// Kita panggil fungsi yang sudah diperbaiki
	checkLimiter, err := s.emailLimiterRepo.GetLimiterEmail(ctx, email)
	if err != nil {
		return ResendEmailVerificationResponse{}, errs.NewInternalServerError(err)
	}

	// LOGIC: Jika checkLimiter TIDAK NIL, artinya Key masih ada di Redis.
	// Berarti user harus menunggu.
	if checkLimiter != nil {
		return ResendEmailVerificationResponse{
			Email:      input.Email,
			RetryAfter: checkLimiter.RetryAfterSeconds,
		}, errs.NewBadRequest("PLEASE_WAIT")
	}

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
		if u.Provider == entity.AuthProviderCredential {
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

	token, err := s.tm.GenerateCreateAccountToken(
		token.GenerateCreateAccountTokenInput{
			ID:       profile.ID,
			Email:    profile.Email,
			Name:     profile.Name,
			ImageUrl: imageUrl,
		},
	)
	if err != nil {
		return ResendEmailVerificationResponse{}, errs.NewInternalServerError(err)
	}

	// Enqueue User Verification
	ctxQ, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	err = s.queue.EnqueueUserVerification(ctxQ, mailer.VerificationInputDTO{
		Name:  profile.Name,
		To:    profile.Email,
		Token: token,
		From:  input.From,
	})
	if err != nil {
		return ResendEmailVerificationResponse{}, err
	}

	//  SET RATE LIMITER BARU
	// Gunakan duration dari Config
	retryAfter := s.cfg.CAN_RESEND_EMAIL_AFTER // misal 60 detik
	retryAfterDuration := time.Duration(retryAfter) * time.Second

	// Simpan ke Redis (Hanya Email & TTL)
	err = s.emailLimiterRepo.SaveLimiterEmail(ctx, profile.Email, retryAfterDuration)
	if err != nil {
		logger.From(ctx).Warn("failed to save email limiter", "err", err)
	}

	return ResendEmailVerificationResponse{
		ID:         profile.ID,
		Name:       profile.Name,
		Email:      profile.Email,
		ImageUrl:   imageUrl,
		RetryAfter: retryAfter,
	}, nil

}
