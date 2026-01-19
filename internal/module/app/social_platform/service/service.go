// internal/module/app/social_platform/service/service.go
package social_platform_service

import (
	"context"
	"database/sql"

	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/pagination"

	"github.com/google/uuid"
)

// SocialPlatformService handles social platform operations
type SocialPlatformService struct {
	store entity.Store
}

// NewService creates a new SocialPlatformService
func NewService(store entity.Store) *SocialPlatformService {
	return &SocialPlatformService{
		store: store,
	}
}

// GetPlatformCodes returns all available platform codes from enum
func (s *SocialPlatformService) GetPlatformCodes() []string {
	return []string{
		"linked_in",
		"facebook_page",
		"instagram_business",
		"whatsapp_business",
		"tiktok",
		"youtube",
		"twitter",
		"pinterest",
	}
}

// GetAll returns paginated social platforms
func (s *SocialPlatformService) GetAll(ctx context.Context, filter GetSocialPlatformsFilter) ([]SocialPlatformResponse, *pagination.Pagination, error) {
	// Count total
	count, err := s.store.CountAllAppSocialPlatforms(ctx, entity.CountAllAppSocialPlatformsParams{
		IncludeInactive: filter.IncludeInactive,
		Search:          filter.Search,
	})
	if err != nil {
		return nil, nil, errs.NewInternalServerError(err)
	}

	pag := pagination.NewPagination(&pagination.PaginationParams{
		Total: int(count),
		Page:  filter.Page,
		Limit: filter.Limit,
	})

	// Calculate offset
	offset := (filter.Page - 1) * filter.Limit
	if offset < 0 {
		offset = 0
	}

	// Get data
	data, err := s.store.GetAllAppSocialPlatforms(ctx, entity.GetAllAppSocialPlatformsParams{
		IncludeInactive: filter.IncludeInactive,
		Search:          filter.Search,
		SortBy:          filter.SortBy,
		SortDir:         filter.SortDir,
		PageLimit:       int32(filter.Limit),
		PageOffset:      int32(offset),
	})
	if err != nil {
		return nil, nil, errs.NewInternalServerError(err)
	}

	responses := make([]SocialPlatformResponse, len(data))
	for i, d := range data {
		responses[i] = mapToResponse(d)
	}

	return responses, &pag, nil
}

// GetById returns social platform by id
func (s *SocialPlatformService) GetById(ctx context.Context, id int64) (SocialPlatformResponse, error) {
	data, err := s.store.GetAppSocialPlatformById(ctx, id)
	if err == sql.ErrNoRows {
		return SocialPlatformResponse{}, errs.NewNotFound("SOCIAL_PLATFORM_NOT_FOUND")
	}
	if err != nil {
		return SocialPlatformResponse{}, errs.NewInternalServerError(err)
	}
	return mapToResponse(data), nil
}

// Create creates new social platform
func (s *SocialPlatformService) Create(ctx context.Context, input CreateSocialPlatformInput) (SocialPlatformResponse, error) {
	// Check duplicate platform_code
	existing, err := s.store.GetAppSocialPlatformByPlatformCode(ctx, entity.SocialPlatformType(input.PlatformCode))
	if err != nil && err != sql.ErrNoRows {
		return SocialPlatformResponse{}, errs.NewInternalServerError(err)
	}
	if existing.ID != 0 {
		return SocialPlatformResponse{}, errs.NewBadRequest("PLATFORM_CODE_ALREADY_EXISTS")
	}

	// Validate platform code is valid enum
	if !isValidPlatformCode(input.PlatformCode) {
		return SocialPlatformResponse{}, errs.NewBadRequest("INVALID_PLATFORM_CODE")
	}

	var result entity.AppSocialPlatform
	err = s.store.ExecTx(ctx, func(q *entity.Queries) error {
		var txErr error
		result, txErr = q.CreateAppSocialPlatform(ctx, entity.CreateAppSocialPlatformParams{
			PlatformCode: entity.SocialPlatformType(input.PlatformCode),
			Logo:         toNullString(input.Logo),
			Name:         input.Name,
			Hint:         input.Hint,
			IsActive:     input.IsActive,
		})
		if txErr != nil {
			return txErr
		}

		// Track change
		_, txErr = q.CreateAppSocialPlatformChange(ctx, entity.CreateAppSocialPlatformChangeParams{
			Action:             entity.ActionChangeTypeCreate,
			ProfileID:          input.ProfileID,
			SocialPlatformID:   result.ID,
			BeforePlatformCode: entity.SocialPlatformType(input.PlatformCode),
			BeforeLogo:         toNullString(input.Logo),
			BeforeName:         input.Name,
			BeforeHint:         input.Hint,
			BeforeIsActive:     input.IsActive,
			AfterPlatformCode:  entity.SocialPlatformType(input.PlatformCode),
			AfterLogo:          toNullString(input.Logo),
			AfterName:          input.Name,
			AfterHint:          input.Hint,
			AfterIsActive:      input.IsActive,
		})
		return txErr
	})
	if err != nil {
		return SocialPlatformResponse{}, errs.NewInternalServerError(err)
	}

	return mapToResponse(result), nil
}

// Update updates social platform
func (s *SocialPlatformService) Update(ctx context.Context, id int64, input UpdateSocialPlatformInput) (SocialPlatformResponse, error) {
	// Get existing
	existing, err := s.store.GetAppSocialPlatformById(ctx, id)
	if err == sql.ErrNoRows {
		return SocialPlatformResponse{}, errs.NewNotFound("SOCIAL_PLATFORM_NOT_FOUND")
	}
	if err != nil {
		return SocialPlatformResponse{}, errs.NewInternalServerError(err)
	}

	// Validate platform code is valid enum
	if !isValidPlatformCode(input.PlatformCode) {
		return SocialPlatformResponse{}, errs.NewBadRequest("INVALID_PLATFORM_CODE")
	}

	// Check duplicate platform_code if changed
	if string(existing.PlatformCode) != input.PlatformCode {
		dup, err := s.store.GetAppSocialPlatformByPlatformCode(ctx, entity.SocialPlatformType(input.PlatformCode))
		if err != nil && err != sql.ErrNoRows {
			return SocialPlatformResponse{}, errs.NewInternalServerError(err)
		}
		if dup.ID != 0 {
			return SocialPlatformResponse{}, errs.NewBadRequest("PLATFORM_CODE_ALREADY_EXISTS")
		}
	}

	var result entity.AppSocialPlatform
	err = s.store.ExecTx(ctx, func(q *entity.Queries) error {
		var txErr error
		result, txErr = q.UpdateAppSocialPlatform(ctx, entity.UpdateAppSocialPlatformParams{
			ID:           id,
			PlatformCode: entity.SocialPlatformType(input.PlatformCode),
			Logo:         toNullString(input.Logo),
			Name:         input.Name,
			Hint:         input.Hint,
			IsActive:     input.IsActive,
		})
		if txErr != nil {
			return txErr
		}

		// Track change
		_, txErr = q.CreateAppSocialPlatformChange(ctx, entity.CreateAppSocialPlatformChangeParams{
			Action:             entity.ActionChangeTypeUpdate,
			ProfileID:          input.ProfileID,
			SocialPlatformID:   id,
			BeforePlatformCode: existing.PlatformCode,
			BeforeLogo:         existing.Logo,
			BeforeName:         existing.Name,
			BeforeHint:         existing.Hint,
			BeforeIsActive:     existing.IsActive,
			AfterPlatformCode:  entity.SocialPlatformType(input.PlatformCode),
			AfterLogo:          toNullString(input.Logo),
			AfterName:          input.Name,
			AfterHint:          input.Hint,
			AfterIsActive:      input.IsActive,
		})
		return txErr
	})
	if err != nil {
		return SocialPlatformResponse{}, errs.NewInternalServerError(err)
	}

	return mapToResponse(result), nil
}

// Delete soft deletes social platform
func (s *SocialPlatformService) Delete(ctx context.Context, id int64, profileID uuid.UUID) (SocialPlatformResponse, error) {
	// Get existing
	existing, err := s.store.GetAppSocialPlatformById(ctx, id)
	if err == sql.ErrNoRows {
		return SocialPlatformResponse{}, errs.NewNotFound("SOCIAL_PLATFORM_NOT_FOUND")
	}
	if err != nil {
		return SocialPlatformResponse{}, errs.NewInternalServerError(err)
	}

	var result entity.AppSocialPlatform
	err = s.store.ExecTx(ctx, func(q *entity.Queries) error {
		var txErr error
		result, txErr = q.DeleteAppSocialPlatform(ctx, id)
		if txErr != nil {
			return txErr
		}

		// Track change
		_, txErr = q.CreateAppSocialPlatformChange(ctx, entity.CreateAppSocialPlatformChangeParams{
			Action:             entity.ActionChangeTypeDelete,
			ProfileID:          profileID,
			SocialPlatformID:   id,
			BeforePlatformCode: existing.PlatformCode,
			BeforeLogo:         existing.Logo,
			BeforeName:         existing.Name,
			BeforeHint:         existing.Hint,
			BeforeIsActive:     existing.IsActive,
			AfterPlatformCode:  existing.PlatformCode,
			AfterLogo:          existing.Logo,
			AfterName:          existing.Name,
			AfterHint:          existing.Hint,
			AfterIsActive:      false,
		})
		return txErr
	})
	if err != nil {
		return SocialPlatformResponse{}, errs.NewInternalServerError(err)
	}

	return mapToResponse(result), nil
}

// mapToResponse maps entity to response
func mapToResponse(p entity.AppSocialPlatform) SocialPlatformResponse {
	var logo *string
	if p.Logo.Valid {
		logo = &p.Logo.String
	}
	return SocialPlatformResponse{
		ID:           p.ID,
		PlatformCode: string(p.PlatformCode),
		Logo:         logo,
		Name:         p.Name,
		Hint:         p.Hint,
		IsActive:     p.IsActive,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

// toNullString converts *string to sql.NullString
func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

// isValidPlatformCode checks if platform code is valid enum value
func isValidPlatformCode(code string) bool {
	validCodes := []string{
		"linked_in",
		"facebook_page",
		"instagram_business",
		"whatsapp_business",
		"tiktok",
		"youtube",
		"twitter",
		"pinterest",
	}
	for _, c := range validCodes {
		if c == code {
			return true
		}
	}
	return false
}
