// internal/module/creator/business_creator_image/service/service.go
package business_creator_image_service

import (
	"context"
	"database/sql"
	"encoding/json"

	creator_image_service "postmatic-api/internal/module/creator/creator_image/service"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/pagination"
	"postmatic-api/pkg/utils"
)

// NotShowingReason constants
const (
	ReasonBanned       = "CONTENT_IMAGE_BANNED"
	ReasonNotPublished = "CONTENT_IMAGE_CURRENTLY_NOT_PUBLISHED"
	ReasonDeleted      = "CONTENT_IMAGE_DELETED"
)

type BusinessCreatorImageService struct {
	store           entity.Store
	creatorImageSvc *creator_image_service.CreatorImageService
}

func NewService(store entity.Store, creatorImageSvc *creator_image_service.CreatorImageService) *BusinessCreatorImageService {
	return &BusinessCreatorImageService{
		store:           store,
		creatorImageSvc: creatorImageSvc,
	}
}

func unmarshalJSONAny(v any, dst any) error {
	if v == nil {
		return nil
	}

	switch t := v.(type) {
	case []byte:
		if len(t) == 0 {
			return nil
		}
		return json.Unmarshal(t, dst)

	case string:
		if t == "" {
			return nil
		}
		return json.Unmarshal([]byte(t), dst)

	case json.RawMessage:
		if len(t) == 0 {
			return nil
		}
		return json.Unmarshal(t, dst)

	default:
		b, err := json.Marshal(t)
		if err != nil {
			return err
		}
		if len(b) == 0 {
			return nil
		}
		return json.Unmarshal(b, dst)
	}
}

func (s *BusinessCreatorImageService) GetSavedCreatorImages(
	ctx context.Context,
	filter GetSavedCreatorImageFilter,
) ([]SavedCreatorImageResponse, *pagination.Pagination, error) {

	params := entity.GetAllSavedCreatorImageByBusinessIdParams{
		BusinessRootID:    filter.BusinessRootID,
		Search:            sql.NullString{String: filter.Search, Valid: filter.Search != ""},
		SortBy:            sql.NullString{String: filter.SortBy, Valid: filter.SortBy != ""},
		SortDir:           sql.NullString{String: filter.SortDir, Valid: filter.SortDir != ""},
		DateStart:         utils.NullStringToNullTime(filter.DateStart),
		DateEnd:           utils.NullStringToNullTime(filter.DateEnd),
		TypeCategoryID:    utils.NullInt64ToNullInt64(filter.TypeCategoryID),
		ProductCategoryID: utils.NullInt64ToNullInt64(filter.ProductCategoryID),
		PageLimit:         int32(filter.PageLimit),
		PageOffset:        int32(filter.PageOffset),
	}

	rows, err := s.store.GetAllSavedCreatorImageByBusinessId(ctx, params)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, errs.NewInternalServerError(err)
	}

	countParams := entity.CountSavedCreatorImageByBusinessIdParams{
		BusinessRootID:    filter.BusinessRootID,
		Search:            params.Search,
		DateStart:         params.DateStart,
		DateEnd:           params.DateEnd,
		TypeCategoryID:    params.TypeCategoryID,
		ProductCategoryID: params.ProductCategoryID,
	}
	total, err := s.store.CountSavedCreatorImageByBusinessId(ctx, countParams)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, errs.NewInternalServerError(err)
	}

	pag := pagination.NewPagination(&pagination.PaginationParams{
		Total: int(total),
		Page:  filter.Page,
		Limit: filter.PageLimit,
	})

	res := make([]SavedCreatorImageResponse, 0, len(rows))
	for _, r := range rows {
		// decode jsonb -> []TypeCategorySub / []ProductCategorySub
		var typeSubs []TypeCategorySub
		var productSubs []ProductCategorySub

		_ = unmarshalJSONAny(r.TypeCategorySubs, &typeSubs)
		_ = unmarshalJSONAny(r.ProductCategorySubs, &productSubs)

		var publisher *PublisherSub
		if r.PublisherID.Valid {
			name := ""
			if r.PublisherName.Valid {
				name = r.PublisherName.String
			}
			var img *string
			if r.PublisherImage.Valid {
				img = &r.PublisherImage.String
			}
			publisher = &PublisherSub{
				ID:    r.PublisherID.UUID.String(),
				Name:  name,
				Image: img,
			}
		}

		// Determine notShowingReason and imageUrl
		var notShowingReason *string
		var imageUrl *string

		// Check in priority order: deleted > banned > not published
		if r.CreatorImageDeletedAt.Valid {
			reason := ReasonDeleted
			notShowingReason = &reason
			imageUrl = nil
		} else if r.IsBanned {
			reason := ReasonBanned
			notShowingReason = &reason
			imageUrl = nil
		} else if !r.IsPublished {
			reason := ReasonNotPublished
			notShowingReason = &reason
			imageUrl = nil
		} else {
			notShowingReason = nil
			imageUrl = &r.ImageUrl
		}

		res = append(res, SavedCreatorImageResponse{
			ID:                  r.CreatorImageID,
			Name:                r.Name,
			ImageURL:            imageUrl,
			IsPublished:         r.IsPublished,
			Price:               r.Price,
			Publisher:           publisher,
			TypeCategorySubs:    typeSubs,
			ProductCategorySubs: productSubs,
			NotShowingReason:    notShowingReason,
			SavedAt:             r.SavedAt,
			CreatedAt:           r.CreatedAt.Time,
			UpdatedAt:           r.UpdatedAt.Time,
		})
	}

	return res, &pag, nil
}

func (s *BusinessCreatorImageService) CreateSavedCreatorImage(
	ctx context.Context,
	input CreateSavedInput,
) (SavedCreatorImageActionResponse, error) {

	// 1. Validate creator image exists and check status
	detail, err := s.creatorImageSvc.GetCreatorImageDetailById(ctx, input.CreatorImageID)
	if err != nil {
		return SavedCreatorImageActionResponse{}, err
	}
	if detail == nil {
		return SavedCreatorImageActionResponse{}, errs.NewNotFound("CREATOR_IMAGE_NOT_FOUND")
	}

	// 2. Check if deleted
	if detail.IsDeleted {
		return SavedCreatorImageActionResponse{}, errs.NewBadRequest("CREATOR_IMAGE_DELETED")
	}

	// 3. Check if banned
	if detail.IsBanned {
		return SavedCreatorImageActionResponse{}, errs.NewBadRequest("CREATOR_IMAGE_BANNED")
	}

	// 4. Check if not published
	if !detail.IsPublished {
		return SavedCreatorImageActionResponse{}, errs.NewBadRequest("CREATOR_IMAGE_NOT_PUBLISHED")
	}

	// 5. Check if already saved
	exists, err := s.store.CheckSavedCreatorImageExists(ctx, entity.CheckSavedCreatorImageExistsParams{
		BusinessRootID: input.BusinessRootID,
		CreatorImageID: input.CreatorImageID,
	})
	if err != nil {
		return SavedCreatorImageActionResponse{}, errs.NewInternalServerError(err)
	}
	if exists {
		return SavedCreatorImageActionResponse{}, errs.NewBadRequest("CREATOR_IMAGE_ALREADY_SAVED")
	}

	// 6. Create saved
	saved, err := s.store.CreateSavedCreatorImage(ctx, entity.CreateSavedCreatorImageParams{
		BusinessRootID: input.BusinessRootID,
		CreatorImageID: input.CreatorImageID,
	})
	if err != nil {
		return SavedCreatorImageActionResponse{}, errs.NewInternalServerError(err)
	}

	return SavedCreatorImageActionResponse{
		ID:             saved.ID,
		BusinessRootID: saved.BusinessRootID,
		CreatorImageID: saved.CreatorImageID,
		CreatedAt:      saved.CreatedAt,
	}, nil
}

func (s *BusinessCreatorImageService) DeleteSavedCreatorImage(
	ctx context.Context,
	input DeleteSavedInput,
) (SavedCreatorImageActionResponse, error) {

	// 1. Check if saved exists
	saved, err := s.store.GetSavedCreatorImageByBusinessAndCreatorImage(ctx, entity.GetSavedCreatorImageByBusinessAndCreatorImageParams{
		BusinessRootID: input.BusinessRootID,
		CreatorImageID: input.CreatorImageID,
	})
	if err == sql.ErrNoRows {
		return SavedCreatorImageActionResponse{}, errs.NewNotFound("SAVED_CREATOR_IMAGE_NOT_FOUND")
	}
	if err != nil {
		return SavedCreatorImageActionResponse{}, errs.NewInternalServerError(err)
	}

	// 2. Soft delete
	err = s.store.SoftDeleteSavedCreatorImage(ctx, entity.SoftDeleteSavedCreatorImageParams{
		BusinessRootID: input.BusinessRootID,
		CreatorImageID: input.CreatorImageID,
	})
	if err != nil {
		return SavedCreatorImageActionResponse{}, errs.NewInternalServerError(err)
	}

	return SavedCreatorImageActionResponse{
		ID:             saved.ID,
		BusinessRootID: saved.BusinessRootID,
		CreatorImageID: saved.CreatorImageID,
		CreatedAt:      saved.CreatedAt,
	}, nil
}
