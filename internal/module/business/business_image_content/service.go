// internal/module/business/business_image_content/service.go
package business_image_content

import (
	"context"
	"database/sql"

	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/pagination"
	"postmatic-api/pkg/utils"
)

type BusinessImageContentService struct {
	store entity.Store
}

func NewService(store entity.Store) *BusinessImageContentService {
	return &BusinessImageContentService{
		store: store,
	}
}

func (s *BusinessImageContentService) GetBusinessImageContentsByBusinessRootID(ctx context.Context, filter GetBusinessImageContentsByBusinessRootIDFilter) ([]BusinessImageContentResponse, pagination.Pagination, error) {

	inputFilter := entity.GetBusinessImageContentsByBusinessRootIdParams{
		BusinessRootID: filter.BusinessRootID,
		Search:         filter.Search,
		SortBy:         string(filter.SortBy),
		PageOffset:     int32(filter.PageOffset),
		PageLimit:      int32(filter.PageLimit),
		SortDir:        string(filter.SortDir),
		DateStart:      utils.NullStringToNullTime(filter.DateStart),
		DateEnd:        utils.NullStringToNullTime(filter.DateEnd),
		ReadyToPost:    utils.NullBoolPtrToNullBool(filter.ReadyToPost),
	}

	bk, err := s.store.GetBusinessImageContentsByBusinessRootId(ctx, inputFilter)
	if err != nil && err != sql.ErrNoRows {
		return []BusinessImageContentResponse{}, pagination.Pagination{}, errs.NewInternalServerError(err)
	}

	if bk == nil {
		return []BusinessImageContentResponse{}, pagination.Pagination{}, nil
	}

	var result []BusinessImageContentResponse
	for _, v := range bk {

		var bProdId *int64
		if v.BusinessProductID.Valid {
			bProdId = &v.BusinessProductID.Int64
		}
		result = append(result, BusinessImageContentResponse{
			BusinessRootID:    filter.BusinessRootID,
			Category:          v.Category,
			ImageUrls:         v.ImageUrls,
			ID:                v.ID,
			Caption:           v.Caption.String,
			Type:              string(v.Type),
			ReadyToPost:       v.ReadyToPost,
			CreatedAt:         v.CreatedAt.Time,
			UpdatedAt:         v.UpdatedAt.Time,
			BusinessProductID: bProdId,
		})
	}

	countParam := entity.CountBusinessImageContentsByBusinessRootIdParams{
		BusinessRootID: filter.BusinessRootID,
		Search:         filter.Search,
		DateStart:      utils.NullStringToNullTime(filter.DateStart),
		DateEnd:        utils.NullStringToNullTime(filter.DateEnd),
		ReadyToPost:    utils.NullBoolPtrToNullBool(filter.ReadyToPost),
	}

	count, err := s.store.CountBusinessImageContentsByBusinessRootId(ctx, countParam)
	if err != nil {
		return []BusinessImageContentResponse{}, pagination.Pagination{}, errs.NewInternalServerError(err)
	}

	paginationParams := pagination.PaginationParams{
		Total: int(count),
		Page:  filter.Page,
		Limit: filter.PageLimit,
	}

	pagination := pagination.NewPagination(&paginationParams)

	return result, pagination, nil
}

func (s *BusinessImageContentService) CreateBusinessImageContent(ctx context.Context, input CreateUpdateBusinessImageContentInput) (*BusinessImageContentResponse, error) {
	inputParam := entity.CreateBusinessImageContentParams{
		BusinessRootID:    input.BusinessRootID,
		Category:          input.Category,
		ImageUrls:         input.ImageUrls,
		Caption:           sql.NullString{String: input.Caption, Valid: input.Caption != ""},
		Type:              entity.BusinessImageContentType(input.Type),
		ReadyToPost:       input.ReadyToPost,
		BusinessProductID: utils.NullInt64ToNullInt64(input.BusinessProductID),
	}

	created, err := s.store.CreateBusinessImageContent(ctx, inputParam)
	if err != nil {
		return nil, errs.NewInternalServerError(err)
	}

	return &BusinessImageContentResponse{
		BusinessRootID:    input.BusinessRootID,
		Category:          created.Category,
		ImageUrls:         created.ImageUrls,
		ID:                created.ID,
		Caption:           created.Caption.String,
		Type:              string(created.Type),
		ReadyToPost:       created.ReadyToPost,
		CreatedAt:         created.CreatedAt.Time,
		UpdatedAt:         created.UpdatedAt.Time,
		BusinessProductID: nil,
	}, nil
}

func (s *BusinessImageContentService) UpdateBusinessImageContent(ctx context.Context, input CreateUpdateBusinessImageContentInput, id int64) (*BusinessImageContentResponse, error) {
	inputParam := entity.UpdateBusinessImageContentParams{
		ID:          id,
		Category:    input.Category,
		ImageUrls:   input.ImageUrls,
		Caption:     sql.NullString{String: input.Caption, Valid: input.Caption != ""},
		Type:        entity.BusinessImageContentType(input.Type),
		ReadyToPost: input.ReadyToPost,
	}

	updated, err := s.store.UpdateBusinessImageContent(ctx, inputParam)

	if err == sql.ErrNoRows || updated.DeletedAt.Valid {
		return nil, errs.NewNotFound("")
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, errs.NewInternalServerError(err)
	}

	return &BusinessImageContentResponse{
		BusinessRootID:    input.BusinessRootID,
		Category:          updated.Category,
		ImageUrls:         updated.ImageUrls,
		ID:                updated.ID,
		Caption:           updated.Caption.String,
		Type:              string(updated.Type),
		ReadyToPost:       updated.ReadyToPost,
		CreatedAt:         updated.CreatedAt.Time,
		UpdatedAt:         updated.UpdatedAt.Time,
		BusinessProductID: nil,
	}, nil
}

func (s *BusinessImageContentService) DeleteBusinessImageContent(ctx context.Context, id int64) (*BusinessImageContentResponse, error) {
	deleted, err := s.store.SoftDeleteBusinessImageContentByBusinessImageContentId(ctx, id)
	if err == sql.ErrNoRows {
		return nil, errs.NewNotFound("")
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, errs.NewInternalServerError(err)
	}

	return &BusinessImageContentResponse{
		BusinessRootID:    deleted.BusinessRootID,
		Category:          deleted.Category,
		ImageUrls:         deleted.ImageUrls,
		ID:                deleted.ID,
		Caption:           deleted.Caption.String,
		Type:              string(deleted.Type),
		ReadyToPost:       deleted.ReadyToPost,
		CreatedAt:         deleted.CreatedAt.Time,
		UpdatedAt:         deleted.UpdatedAt.Time,
		BusinessProductID: nil,
	}, nil
}
