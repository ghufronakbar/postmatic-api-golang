// internal/module/app/category_creator_image/service.go
package category_creator_image

import (
	"context"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/pagination"
)

type CategoryCreatorImageService struct {
	store entity.Store
}

func NewCategoryCreatorImageService(store entity.Store) *CategoryCreatorImageService {
	return &CategoryCreatorImageService{store: store}
}

func (s *CategoryCreatorImageService) GetCategoryCreatorImageType(ctx context.Context, filter GetCategoryCreatorImageTypeFilter) ([]CategoryCreatorImageTypeResponse, *pagination.Pagination, error) {

	filterData := entity.GetAllAppCreatorImageTypeCategoriesParams{
		Search:     filter.Search,
		SortBy:     filter.SortBy,
		SortDir:    filter.SortDir,
		PageOffset: int32(filter.PageOffset),
		PageLimit:  int32(filter.PageLimit),
	}

	categories, err := s.store.GetAllAppCreatorImageTypeCategories(ctx, filterData)
	if err != nil {
		return nil, nil, err
	}

	count, err := s.store.CountAllAppCreatorImageTypeCategories(ctx, filterData.Search)
	if err != nil {
		return nil, nil, err
	}

	paginationParams := pagination.PaginationParams{
		Total: int(count),
		Page:  filter.Page,
		Limit: filter.PageLimit,
	}

	pagination := pagination.NewPagination(&paginationParams)

	var responses []CategoryCreatorImageTypeResponse
	for _, category := range categories {
		responses = append(responses, CategoryCreatorImageTypeResponse{
			ID:        category.ID,
			Name:      category.Name,
			TotalData: category.TotalData,
		})
	}
	return responses, &pagination, nil
}

func (s *CategoryCreatorImageService) GetCategoryCreatorImageProduct(ctx context.Context, filter GetCategoryCreatorImageProductFilter) ([]CategoryCreatorImageProductResponse, *pagination.Pagination, error) {

	filterData := entity.GetAllAppCreatorImageProductCategoriesParams{
		Search:     filter.Search,
		SortBy:     filter.SortBy,
		SortDir:    filter.SortDir,
		PageOffset: int32(filter.PageOffset),
		PageLimit:  int32(filter.PageLimit),
		Locale:     filter.Locale,
	}

	categories, err := s.store.GetAllAppCreatorImageProductCategories(ctx, filterData)
	if err != nil {
		return nil, nil, err
	}

	count, err := s.store.CountAllAppCreatorImageProductCategories(ctx, filterData.Search)
	if err != nil {
		return nil, nil, err
	}

	paginationParams := pagination.PaginationParams{
		Total: int(count),
		Page:  filter.Page,
		Limit: filter.PageLimit,
	}

	pagination := pagination.NewPagination(&paginationParams)

	var responses []CategoryCreatorImageProductResponse
	for _, category := range categories {
		var name string
		if filter.Locale == "id" {
			name = category.IndonesianName
		} else {
			name = category.EnglishName
		}
		responses = append(responses, CategoryCreatorImageProductResponse{
			ID:        category.ID,
			Name:      name,
			TotalData: category.TotalData,
		})
	}
	return responses, &pagination, nil
}
