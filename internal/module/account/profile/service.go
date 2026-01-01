// internal/module/account/profile/service.go
package profile

import (
	"context"
	"database/sql"
	"time"

	"postmatic-api/config"
	"postmatic-api/internal/module/headless/mailer"
	"postmatic-api/internal/module/headless/queue"
	"postmatic-api/internal/module/headless/token"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/logger"

	"postmatic-api/pkg/utils"

	emailLimiterRepo "postmatic-api/internal/repository/redis/email_limiter_repository"

	"github.com/google/uuid"
)

type ProfileService struct {
	store            entity.Store
	queue            queue.MailerProducer
	cfg              config.Config
	emailLimiterRepo *emailLimiterRepo.LimiterEmailRepo
	tm               token.TokenMaker
}

// Update Constructor: Minta Token Maker dari main.go
func NewService(store entity.Store, queue queue.MailerProducer, cfg config.Config, emailLimiterRepo *emailLimiterRepo.LimiterEmailRepo, tm token.TokenMaker) *ProfileService {
	return &ProfileService{
		store:            store,
		queue:            queue,
		cfg:              cfg,
		emailLimiterRepo: emailLimiterRepo,
		tm:               tm,
	}
}

func (s *ProfileService) GetProfile(ctx context.Context, profileId uuid.UUID) (GetProfileResponse, error) {
	profile, err := s.store.GetProfileById(ctx, profileId)

	if err == sql.ErrNoRows {
		return GetProfileResponse{}, errs.NewUnauthorized("PROFILE_NOT_FOUND")
	}

	if err != nil {
		return GetProfileResponse{}, err
	}

	users, err := s.store.GetUserByEmailProfile(ctx, profile.Email)
	if err != nil {
		return GetProfileResponse{}, err
	}

	credUser := CredentialUser{
		IsPasswordSet: false,
		VerifiedAt:    nil,
	}
	googleUser := GoogleUser{
		IsConnected: false,
		VerifiedAt:  nil,
	}

	for _, user := range users {
		// Check is credential connected
		if user.Provider == entity.AuthProviderCredential {
			isPasswordSet := false
			var verifiedAt *time.Time
			if user.Password.Valid {
				isPasswordSet = true
			}
			if user.VerifiedAt.Valid {
				verifiedAt = &user.VerifiedAt.Time
			}
			credUser = CredentialUser{
				IsPasswordSet: isPasswordSet,
				VerifiedAt:    verifiedAt,
			}
		}
		// Check is google connected
		if user.Provider == entity.AuthProviderGoogle {
			var verifiedAt *time.Time
			if user.VerifiedAt.Valid {
				verifiedAt = &user.VerifiedAt.Time
			}
			googleUser = GoogleUser{
				IsConnected: true,
				VerifiedAt:  verifiedAt,
			}
		}
	}

	profileResponse := GetProfileResponse{
		ID:          profile.ID,
		Name:        profile.Name,
		Email:       profile.Email,
		ImageUrl:    utils.NullStringToString(profile.ImageUrl),
		CountryCode: utils.NullStringToStringVal(profile.CountryCode),
		Phone:       utils.NullStringToStringVal(profile.Phone),
		Description: utils.NullStringToString(profile.Description),
		Credential:  credUser,
		Google:      googleUser,
	}
	return profileResponse, nil
}

func (s *ProfileService) UpdateProfile(ctx context.Context, profileId uuid.UUID, input UpdateProfileInput) (UpdateProfileResponse, error) {
	profile, err := s.GetProfile(ctx, profileId)
	if err != nil {
		return UpdateProfileResponse{}, err
	}

	var imageUrl string
	if input.ImageUrl != nil {
		imageUrl = *input.ImageUrl
	}

	profileUpdated, err := s.store.UpdateProfile(ctx, entity.UpdateProfileParams{
		ID:          profileId,
		Name:        input.Name,
		ImageUrl:    sql.NullString{String: imageUrl, Valid: imageUrl != ""},
		CountryCode: sql.NullString{String: input.CountryCode, Valid: input.CountryCode != ""},
		Phone:       sql.NullString{String: *input.Phone, Valid: input.Phone != nil},
		Description: sql.NullString{String: *input.Description, Valid: input.Description != nil},
	})

	if err != nil {
		return UpdateProfileResponse{}, err
	}

	profileResponse := GetProfileResponse{
		ID:          profileUpdated.ID,
		Name:        profileUpdated.Name,
		Email:       profileUpdated.Email,
		Credential:  profile.Credential,
		Google:      profile.Google,
		ImageUrl:    utils.NullStringToString(profileUpdated.ImageUrl),
		CountryCode: utils.NullStringToStringVal(profileUpdated.CountryCode),
		Phone:       utils.NullStringToStringVal(profileUpdated.Phone),
		Description: utils.NullStringToString(profileUpdated.Description),
		CreatedAt:   profile.CreatedAt,
		UpdatedAt:   profile.UpdatedAt,
	}

	accessToken, err := s.tm.GenerateAccessToken(
		token.GenerateAccessTokenInput{
			ID:       profileUpdated.ID,
			Email:    profileUpdated.Email,
			Name:     profileUpdated.Name,
			ImageUrl: &imageUrl,
		},
	)
	if err != nil {
		return UpdateProfileResponse{}, err
	}

	updateProfileResponse := UpdateProfileResponse{
		AccessToken:        accessToken,
		GetProfileResponse: profileResponse,
	}

	return updateProfileResponse, nil
}

// untuk pengguna yang login credential  (udah setup credential account)
func (s *ProfileService) UpdatePassword(ctx context.Context, profileId uuid.UUID, input UpdatePasswordInput) error {
	profile, err := s.GetProfile(ctx, profileId)
	if err != nil {
		return err
	}
	if !profile.Credential.IsPasswordSet {
		return errs.NewUnauthorized("CREDENTIAL_USER_NOT_FOUND")
	}

	if input.NewPassword == input.OldPassword {
		return errs.NewBadRequest("PASSWORD_SAME")
	}

	users, err := s.store.GetUserByEmailProfile(ctx, profile.Email)
	if err != nil {
		return err
	}

	var credUserPassword *string
	var credUserId uuid.UUID
	for _, user := range users {
		if user.Provider == entity.AuthProviderCredential && user.Password.Valid {
			credUserPassword = &user.Password.String
			credUserId = user.ID
		}
	}

	if credUserPassword == nil {
		return errs.NewBadRequest("CREDENTIAL_USER_NOT_FOUND")
	}

	if credUserId == uuid.Nil {
		return errs.NewBadRequest("CREDENTIAL_USER_NOT_FOUND")
	}

	if !utils.ComparePassword(*credUserPassword, input.OldPassword) {
		return errs.NewBadRequest("PASSWORD_WRONG")
	}

	newHashPassword, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return err
	}

	_, err = s.store.UpdateUserPassword(ctx, entity.UpdateUserPasswordParams{
		ID:       credUserId,
		Password: sql.NullString{String: newHashPassword, Valid: true},
	})
	if err != nil {
		return err
	}

	return nil
}

// untuk pengguna yang login oauth, lalu ingin setup password credential account
// PERBAIKAN LOGIC SETUP PASSWORD
func (s *ProfileService) SetupPassword(ctx context.Context, profileId uuid.UUID, input SetupPasswordInput) (SetupPasswordResponse, error) {

	profile, err := s.GetProfile(ctx, profileId)
	if err != nil {
		return SetupPasswordResponse{}, err
	}

	// Cek apakah user credential sudah ada/password sudah di-set
	if profile.Credential.IsPasswordSet {
		return SetupPasswordResponse{}, errs.NewBadRequest("PASSWORD_ALREADY_SET")
	}

	// 1. CEK LIMITER (Gunakan Email agar konsisten dengan Register)
	checkLimiter, _ := s.emailLimiterRepo.GetLimiterEmail(ctx, profile.Email)
	if checkLimiter != nil {
		return SetupPasswordResponse{
			RetryAfter: checkLimiter.RetryAfterSeconds,
		}, errs.NewBadRequest("PLEASE_WAIT")
	}

	newHashPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return SetupPasswordResponse{}, err
	}

	retryAfter := s.cfg.CAN_RESEND_EMAIL_AFTER
	retryAfterDuration := time.Duration(retryAfter) * time.Second

	// Siapkan Data User Baru
	inputUser := entity.CreateUserParams{
		Password:  sql.NullString{String: newHashPassword, Valid: true},
		ProfileID: profileId,
		Provider:  entity.AuthProviderCredential, // Pastikan pakai ENUM atau Const
	}

	e := s.store.ExecTx(ctx, func(q *entity.Queries) error {
		_, err := s.store.CreateUser(ctx, inputUser)
		if err != nil {
			return err
		}

		// 2. GENERATE JWT TOKEN (BUKAN UUID)
		// Kita butuh token untuk verifikasi email, sama seperti saat register
		createAccountToken, err := s.tm.GenerateCreateAccountToken(
			token.GenerateCreateAccountTokenInput{
				ID:       profile.ID, // ID Profile UUID
				Email:    profile.Email,
				Name:     profile.Name,
				ImageUrl: profile.ImageUrl,
			},
		)
		if err != nil {
			return err
		}

		// 3. KIRIM EMAIL (Kirim JWT Token, bukan User ID)
		// TODO: add to queue instead synchronous (and place outside db transaction)
		ctxQ, cancelQ := context.WithTimeout(ctx, 5*time.Second)
		defer cancelQ()
		err = s.queue.EnqueueUserVerification(ctxQ, mailer.VerificationInputDTO{
			Name:  profile.Name,
			To:    profile.Email,
			Token: createAccountToken,
			From:  input.From,
		})
		if err != nil {
			logger.From(ctx).Error("Failed to enqueue user verification", "error", err)
		}

		// 4. Simpan Limiter (Gunakan Email)
		err = s.emailLimiterRepo.SaveLimiterEmail(ctx, profile.Email, retryAfterDuration)
		if err != nil {
			return err
		}
		return nil
	})

	if e != nil {
		return SetupPasswordResponse{}, e
	}

	return SetupPasswordResponse{
		RetryAfter: retryAfter,
	}, nil
}
