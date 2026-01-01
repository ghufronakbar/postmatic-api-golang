// internal/module/creator/creator_image/service.go
package creator_image

import (
	"context"
	"database/sql"
	"encoding/json"

	"postmatic-api/internal/module/app/category_creator_image"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/pagination"
	"postmatic-api/pkg/utils"

	"github.com/google/uuid"
)

type CreatorImageService struct {
	store entity.Store
	cat   *category_creator_image.CategoryCreatorImageService
}

func NewService(store entity.Store, cat *category_creator_image.CategoryCreatorImageService) *CreatorImageService {
	return &CreatorImageService{
		store: store,
		cat:   cat,
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
		// fallback: kalau driver ngasih map/slice (bukan bytes), marshal ulang
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

func (s *CreatorImageService) GetCreatorImageByProfileId(
	ctx context.Context,
	filter GetCreatorImageFilter,
) ([]CreatorImageResponse, *pagination.Pagination, error) {

	var profileUUID uuid.UUID
	if filter.ProfileID != "" {
		pUUID, _ := uuid.Parse(filter.ProfileID)
		profileUUID = pUUID
	}

	params := entity.GetAllCreatorImageParams{
		ProfileID:         uuid.NullUUID{UUID: profileUUID, Valid: profileUUID != uuid.Nil},
		Search:            sql.NullString{String: filter.Search, Valid: filter.Search != ""},
		SortBy:            sql.NullString{String: filter.SortBy, Valid: filter.SortBy != ""},
		SortDir:           sql.NullString{String: filter.SortDir, Valid: filter.SortDir != ""},
		DateStart:         utils.NullStringToNullTime(filter.DateStart),
		DateEnd:           utils.NullStringToNullTime(filter.DateEnd),
		TypeCategoryID:    utils.NullInt64ToNullInt64(filter.TypeCategoryID),
		ProductCategoryID: utils.NullInt64ToNullInt64(filter.ProductCategoryID),
		PageLimit:         int32(filter.PageLimit),
		PageOffset:        int32(filter.PageOffset),
		Published:         utils.NullBoolPtrToNullBool(filter.Published),
	}

	rows, err := s.store.GetAllCreatorImage(ctx, params)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, errs.NewInternalServerError(err)
	}

	countParams := entity.CountAllCreatorImageParams{
		ProfileID:         params.ProfileID,
		Search:            params.Search,
		DateStart:         params.DateStart,
		DateEnd:           params.DateEnd,
		TypeCategoryID:    params.TypeCategoryID,
		ProductCategoryID: params.ProductCategoryID,
		Published:         params.Published,
	}
	total, err := s.store.CountAllCreatorImage(ctx, countParams)
	if err != nil && err != sql.ErrNoRows {
		return nil, nil, errs.NewInternalServerError(err)
	}

	pag := pagination.NewPagination(&pagination.PaginationParams{
		Total: int(total),
		Page:  filter.Page,
		Limit: filter.PageLimit,
	})

	res := make([]CreatorImageResponse, 0, len(rows))
	for _, r := range rows {
		// decode jsonb -> []TypeCategorySub / []ProductCategorySub
		var typeSubs []TypeCategorySub
		var productSubs []ProductCategorySub

		_ = unmarshalJSONAny(r.TypeCategorySubs, &typeSubs)
		_ = unmarshalJSONAny(r.ProductCategorySubs, &productSubs)

		var publisher *PublisherSub
		if r.PublisherID.Valid {
			// name/image biasanya nullable karena LEFT JOIN
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

		res = append(res, CreatorImageResponse{
			ID:                  r.ID,
			Name:                r.Name,
			ImageURL:            r.ImageUrl,
			IsPublished:         r.IsPublished,
			Price:               r.Price,
			Publisher:           publisher,
			TypeCategorySubs:    typeSubs,
			ProductCategorySubs: productSubs,
			CreatedAt:           r.CreatedAt.Time,
			UpdatedAt:           r.UpdatedAt.Time,
		})
	}

	return res, &pag, nil
}

func (s *CreatorImageService) CreateCreatorImage(ctx context.Context, input CreateCreatorImageInput) (CreatorImageCreateUpdateDeleteResponse, error) {

	var profileUUID uuid.UUID
	if input.ProfileID != "" {
		pUUID, err := uuid.Parse(input.ProfileID)
		if err != nil {
			return CreatorImageCreateUpdateDeleteResponse{}, errs.NewBadRequest("INVALID_PROFILE_ID")
		}
		profileUUID = pUUID
	}

	checkTypeCategoryIds, err := s.cat.GetCategoryCreatorImageTypeByIds(ctx, input.TypeCategoryIds)
	if err != nil {
		return CreatorImageCreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	if len(checkTypeCategoryIds) != len(input.TypeCategoryIds) {
		return CreatorImageCreateUpdateDeleteResponse{}, errs.NewBadRequest("SOME_TYPE_CATEGORY_IDS_NOT_FOUND")
	}

	checkProductCategoryIds, err := s.cat.GetCategoryCreatorImageProductByIds(ctx, input.ProductCategoryIds)
	if err != nil {
		return CreatorImageCreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	if len(checkProductCategoryIds) != len(input.ProductCategoryIds) {
		return CreatorImageCreateUpdateDeleteResponse{}, errs.NewBadRequest("SOME_PRODUCT_CATEGORY_IDS_NOT_FOUND")
	}

	creatorImageDb, err := s.store.CreateCreatorImage(ctx, entity.CreateCreatorImageParams{
		Name:               input.Name,
		ImageUrl:           input.ImageURL,
		IsPublished:        input.IsPublished,
		TypeCategoryIds:    input.TypeCategoryIds,
		ProductCategoryIds: input.ProductCategoryIds,
		Price:              input.Price,
		ProfileID:          uuid.NullUUID{UUID: profileUUID, Valid: profileUUID != uuid.Nil},
	})

	if err != nil {
		return CreatorImageCreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	var profileIdResponse string
	if profileUUID != uuid.Nil {
		profileIdResponse = profileUUID.String()
	}

	return CreatorImageCreateUpdateDeleteResponse{
		ID:          creatorImageDb.ID,
		Name:        creatorImageDb.Name,
		ImageURL:    creatorImageDb.ImageUrl,
		IsPublished: creatorImageDb.IsPublished,
		ProfileId:   &profileIdResponse,
		Price:       creatorImageDb.Price,
		CreatedAt:   creatorImageDb.CreatedAt.Time,
		UpdatedAt:   creatorImageDb.UpdatedAt.Time,
	}, nil

}

func (s *CreatorImageService) UpdateCreatorImage(ctx context.Context, input UpdateCreatorImageInput) (CreatorImageCreateUpdateDeleteResponse, error) {
	var profileUUID uuid.UUID
	var profileIdResponse *string
	if input.ProfileID != "" {
		pUUID, err := uuid.Parse(input.ProfileID)
		if err != nil {
			return CreatorImageCreateUpdateDeleteResponse{}, errs.NewBadRequest("INVALID_PROFILE_ID")
		}
		profileUUID = pUUID
		profileIdResponse = &input.ProfileID
	}

	checkCreatorImage, err := s.store.GetCreatorImageById(ctx, input.CreatorImageId)
	if err != nil && err != sql.ErrNoRows {
		return CreatorImageCreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	if checkCreatorImage.ID == 0 {
		return CreatorImageCreateUpdateDeleteResponse{}, errs.NewBadRequest("CREATOR_IMAGE_NOT_FOUND")
	}

	if checkCreatorImage.ProfileID.UUID != profileUUID {
		return CreatorImageCreateUpdateDeleteResponse{}, errs.NewForbidden("")
	}

	checkTypeCategoryIds, err := s.cat.GetCategoryCreatorImageTypeByIds(ctx, input.TypeCategoryIds)
	if err != nil {
		return CreatorImageCreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	if len(checkTypeCategoryIds) != len(input.TypeCategoryIds) {
		return CreatorImageCreateUpdateDeleteResponse{}, errs.NewBadRequest("SOME_TYPE_CATEGORY_IDS_NOT_FOUND")
	}

	checkProductCategoryIds, err := s.cat.GetCategoryCreatorImageProductByIds(ctx, input.ProductCategoryIds)
	if err != nil {
		return CreatorImageCreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	if len(checkProductCategoryIds) != len(input.ProductCategoryIds) {
		return CreatorImageCreateUpdateDeleteResponse{}, errs.NewBadRequest("SOME_PRODUCT_CATEGORY_IDS_NOT_FOUND")
	}

	if input.ProfileID != "" {
		profileUUID, err = uuid.Parse(input.ProfileID)
		if err != nil {
			return CreatorImageCreateUpdateDeleteResponse{}, errs.NewBadRequest("INVALID_PROFILE_ID")
		}
	}

	creatorImageDb, err := s.store.UpdateCreatorImage(ctx, entity.UpdateCreatorImageParams{
		ID:                 input.CreatorImageId,
		Name:               input.Name,
		ImageUrl:           input.ImageURL,
		IsPublished:        input.IsPublished,
		TypeCategoryIds:    input.TypeCategoryIds,
		ProductCategoryIds: input.ProductCategoryIds,
		Price:              input.Price,
		ProfileID:          uuid.NullUUID{UUID: profileUUID, Valid: profileUUID != uuid.Nil},
	})
	if err != nil {
		return CreatorImageCreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	return CreatorImageCreateUpdateDeleteResponse{
		ID:          creatorImageDb.ID,
		Name:        creatorImageDb.Name,
		ImageURL:    creatorImageDb.ImageUrl,
		IsPublished: creatorImageDb.IsPublished,
		ProfileId:   profileIdResponse,
		Price:       creatorImageDb.Price,
		CreatedAt:   creatorImageDb.CreatedAt.Time,
		UpdatedAt:   creatorImageDb.UpdatedAt.Time,
	}, nil
}

func (s *CreatorImageService) SoftDeleteCreatorImage(ctx context.Context, creatorImageId int64, profileId string) (CreatorImageCreateUpdateDeleteResponse, error) {
	var profileUUID uuid.UUID
	var profileIdResponse *string
	if profileId != "" {
		pUUID, err := uuid.Parse(profileId)
		if err != nil {
			return CreatorImageCreateUpdateDeleteResponse{}, errs.NewBadRequest("INVALID_PROFILE_ID")
		}
		profileUUID = pUUID
		profileIdResponse = &profileId
	}

	checkCreatorImage, err := s.store.GetCreatorImageById(ctx, creatorImageId)
	if err != nil && err != sql.ErrNoRows {
		return CreatorImageCreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	if checkCreatorImage.ID == 0 || checkCreatorImage.DeletedAt.Valid {
		return CreatorImageCreateUpdateDeleteResponse{}, errs.NewNotFound("")
	}

	if checkCreatorImage.ProfileID.UUID != profileUUID {
		return CreatorImageCreateUpdateDeleteResponse{}, errs.NewForbidden("")
	}

	err = s.store.SoftDeleteCreatorImage(ctx, creatorImageId)
	if err != nil {
		return CreatorImageCreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	return CreatorImageCreateUpdateDeleteResponse{
		ID:          creatorImageId,
		Name:        checkCreatorImage.Name,
		ImageURL:    checkCreatorImage.ImageUrl,
		IsPublished: checkCreatorImage.IsPublished,
		ProfileId:   profileIdResponse,
		Price:       checkCreatorImage.Price,
		CreatedAt:   checkCreatorImage.CreatedAt.Time,
		UpdatedAt:   checkCreatorImage.UpdatedAt.Time,
	}, nil
}
