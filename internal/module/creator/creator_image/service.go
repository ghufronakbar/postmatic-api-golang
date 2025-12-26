// internal/module/creator/creator_image/service.go
package creator_image

import (
	"context"
	"database/sql"

	"postmatic-api/internal/module/app/category_creator_image"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"

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

func (s *CreatorImageService) CreateCreatorImage(ctx context.Context, input CreateUpdateCreatorImageInput, profileId string) (CreatorImageCreateUpdateDeleteResponse, error) {
	var profileUUID uuid.UUID
	var profileIdResponse *string
	if profileId != "" {
		profileIdResponse = &profileId
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

	if profileId != "" {
		profileUUID, err = uuid.Parse(profileId)
		if err != nil {
			return CreatorImageCreateUpdateDeleteResponse{}, errs.NewBadRequest("INVALID_PROFILE_ID")
		}
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

func (s *CreatorImageService) UpdateCreatorImage(ctx context.Context, input CreateUpdateCreatorImageInput, creatorImageId int64, profileId string) (CreatorImageCreateUpdateDeleteResponse, error) {
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

	if profileId != "" {
		profileUUID, err = uuid.Parse(profileId)
		if err != nil {
			return CreatorImageCreateUpdateDeleteResponse{}, errs.NewBadRequest("INVALID_PROFILE_ID")
		}
	}

	creatorImageDb, err := s.store.UpdateCreatorImage(ctx, entity.UpdateCreatorImageParams{
		ID:                 creatorImageId,
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
