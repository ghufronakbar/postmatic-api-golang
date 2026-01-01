// internal/module/business/business_product/service.go
package business_product

import (
	"context"
	"database/sql"

	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/pagination"
	"postmatic-api/pkg/utils"
)

type BusinessProductService struct {
	store entity.Store
}

// Update Constructor: Minta Token Maker dari main.go
func NewService(store entity.Store) *BusinessProductService {
	return &BusinessProductService{
		store: store,
	}
}

func (s *BusinessProductService) GetBusinessProductsByBusinessRootID(ctx context.Context, filter GetBusinessProductsByBusinessRootIDFilter) ([]BusinessProductResponse, pagination.Pagination, error) {
	inputFilter := entity.GetBusinessProductsByBusinessRootIdParams{
		BusinessRootID: filter.BusinessRootID,
		Search:         filter.Search,
		SortBy:         string(filter.SortBy),
		PageOffset:     int32(filter.PageOffset),
		PageLimit:      int32(filter.PageLimit),
		SortDir:        string(filter.SortDir),
		Category:       sql.NullString{String: filter.Category, Valid: filter.Category != ""},
		DateStart:      utils.NullStringToNullTime(filter.DateStart),
		DateEnd:        utils.NullStringToNullTime(filter.DateEnd),
	}

	bk, err := s.store.GetBusinessProductsByBusinessRootId(ctx, inputFilter)
	if err != nil && err != sql.ErrNoRows {
		return []BusinessProductResponse{}, pagination.Pagination{}, errs.NewInternalServerError(err)
	}

	if bk == nil {
		return []BusinessProductResponse{}, pagination.Pagination{}, nil
	}

	var result []BusinessProductResponse
	for _, v := range bk {
		result = append(result, BusinessProductResponse{
			BusinessRootID: filter.BusinessRootID,
			Name:           v.Name,
			Category:       v.Category,
			Description:    v.Description.String,
			Price:          v.Price,
			Currency:       v.Currency,
			ImageUrls:      v.ImageUrls,
			CreatedAt:      v.CreatedAt,
			UpdatedAt:      v.UpdatedAt,
			ID:             v.ID,
		})
	}

	countParam := entity.CountBusinessProductsByBusinessRootIdParams{
		BusinessRootID: filter.BusinessRootID,
		Search:         filter.Search,
		Category:       sql.NullString{String: filter.Category, Valid: filter.Category != ""},
		DateStart:      utils.NullStringToNullTime(filter.DateStart),
		DateEnd:        utils.NullStringToNullTime(filter.DateEnd),
	}

	count, err := s.store.CountBusinessProductsByBusinessRootId(ctx, countParam)
	if err != nil {
		return []BusinessProductResponse{}, pagination.Pagination{}, errs.NewInternalServerError(err)
	}

	paginationParams := pagination.PaginationParams{
		Total: int(count),
		Page:  filter.Page,
		Limit: filter.PageLimit,
	}

	pagination := pagination.NewPagination(&paginationParams)

	return result, pagination, nil
}

func (s *BusinessProductService) CreateBusinessProduct(ctx context.Context, input CreateBusinessProductInput) (BusinessProductResponse, error) {

	inputFilter := entity.CreateBusinessProductParams{
		BusinessRootID: input.BusinessRootID,
		Name:           input.Name,
		Category:       input.Category,
		Description:    sql.NullString{String: input.Description, Valid: input.Description != ""},
		Price:          input.Price,
		Currency:       input.Currency,
		ImageUrls:      input.ImageUrls,
	}

	bk, err := s.store.CreateBusinessProduct(ctx, inputFilter)
	if err != nil {
		return BusinessProductResponse{}, errs.NewInternalServerError(err)
	}

	return BusinessProductResponse{
		BusinessRootID: input.BusinessRootID,
		Name:           bk.Name,
		Category:       bk.Category,
		Description:    bk.Description.String,
		Price:          bk.Price,
		Currency:       bk.Currency,
		ImageUrls:      bk.ImageUrls,
		CreatedAt:      bk.CreatedAt,
		UpdatedAt:      bk.UpdatedAt,
		ID:             bk.ID,
	}, nil
}

func (s *BusinessProductService) UpdateBusinessProduct(ctx context.Context, input UpdateBusinessProductInput) (BusinessProductResponse, error) {

	inputFilter := entity.UpdateBusinessProductParams{
		Name:        input.Name,
		Category:    input.Category,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Price:       input.Price,
		Currency:    input.Currency,
		ImageUrls:   input.ImageUrls,
		ID:          input.ID,
	}

	bk, err := s.store.UpdateBusinessProduct(ctx, inputFilter)
	if err != nil {
		return BusinessProductResponse{}, errs.NewInternalServerError(err)
	}

	return BusinessProductResponse{
		BusinessRootID: bk.BusinessRootID,
		Name:           bk.Name,
		Category:       bk.Category,
		Description:    bk.Description.String,
		Price:          bk.Price,
		Currency:       bk.Currency,
		ImageUrls:      bk.ImageUrls,
		CreatedAt:      bk.CreatedAt,
		UpdatedAt:      bk.UpdatedAt,
		ID:             bk.ID,
	}, nil
}

func (s *BusinessProductService) SoftDeleteBusinessProductByBusinessRootID(ctx context.Context, businessProductId int64) (SoftDeleteBusinessProductResponse, error) {

	check, err := s.store.GetBusinessProductByBusinessProductId(ctx, businessProductId)
	if err == sql.ErrNoRows {
		return SoftDeleteBusinessProductResponse{}, errs.NewNotFound("")
	}
	if err != nil && err != sql.ErrNoRows {
		return SoftDeleteBusinessProductResponse{}, errs.NewInternalServerError(err)
	}
	if check.DeletedAt.Valid {
		return SoftDeleteBusinessProductResponse{}, errs.NewNotFound("")
	}

	bk, err := s.store.SoftDeleteBusinessProductByBusinessProductId(ctx, businessProductId)
	if err != nil {
		return SoftDeleteBusinessProductResponse{}, errs.NewInternalServerError(err)
	}

	return SoftDeleteBusinessProductResponse{
		ID: bk,
	}, nil
}
