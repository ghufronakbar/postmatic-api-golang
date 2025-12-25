// internal/module/app/rss/service.go
package rss

import (
	"context"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/pagination"

	"github.com/google/uuid"
)

type RSSService struct {
	store entity.Store
}

func NewRSSService(store entity.Store) *RSSService {
	return &RSSService{store: store}
}

func (s *RSSService) GetRSSCategory(ctx context.Context, filter GetRSSCategoryFilter) ([]RSSCategoryResponse, *pagination.Pagination, error) {

	filterData := entity.GetAllRSSCategoryParams{
		Search:     filter.Search,
		SortBy:     filter.SortBy,
		SortDir:    filter.SortDir,
		PageOffset: int32(filter.PageOffset),
		PageLimit:  int32(filter.PageLimit),
	}

	categories, err := s.store.GetAllRSSCategory(ctx, filterData)
	if err != nil {
		return nil, nil, err
	}

	count, err := s.store.CountAllRSSCategory(ctx, filter.Search)
	if err != nil {
		return nil, nil, err
	}

	paginationParams := pagination.PaginationParams{
		Total: int(count),
		Page:  filter.Page,
		Limit: filter.PageLimit,
	}

	pagination := pagination.NewPagination(&paginationParams)

	var responses []RSSCategoryResponse
	for _, category := range categories {
		responses = append(responses, RSSCategoryResponse{
			ID:        category.ID.String(),
			Name:      category.Name,
			CreatedAt: category.CreatedAt,
			UpdatedAt: category.UpdatedAt,
		})
	}
	return responses, &pagination, nil
}

func (s *RSSService) GetRSSFeed(ctx context.Context, filter GetRSSFeedFilter) ([]RSSResponse, *pagination.Pagination, error) {

	var categoryUUID uuid.UUID
	if filter.Category != "" {
		cUUID, err := uuid.Parse(filter.Category)
		if err != nil {
			categoryUUID = uuid.Nil
		} else {
			categoryUUID = cUUID
		}
	}

	filterData := entity.GetAllRSSFeedParams{
		Search:     filter.Search,
		SortBy:     filter.SortBy,
		SortDir:    filter.SortDir,
		PageOffset: int32(filter.PageOffset),
		PageLimit:  int32(filter.PageLimit),
		Category:   uuid.NullUUID{UUID: categoryUUID, Valid: filter.Category != uuid.Nil.String()},
	}
	feeds, err := s.store.GetAllRSSFeed(ctx, filterData)
	if err != nil {
		return nil, nil, err
	}
	filterCount := entity.CountAllRSSFeedParams{
		Search:   filter.Search,
		Category: uuid.NullUUID{UUID: categoryUUID, Valid: filter.Category != uuid.Nil.String()},
	}
	count, err := s.store.CountAllRSSFeed(ctx, filterCount)
	if err != nil {
		return nil, nil, err
	}

	paginationParams := pagination.PaginationParams{
		Total: int(count),
		Page:  filter.Page,
		Limit: filter.PageLimit,
	}

	pagination := pagination.NewPagination(&paginationParams)
	var responses []RSSResponse
	for _, feed := range feeds {
		responses = append(responses, RSSResponse{
			ID:                  feed.ID.String(),
			Title:               feed.Title,
			URL:                 feed.Url,
			Publisher:           feed.Publisher,
			MasterRSSCategoryID: feed.AppRssCategoryID.String(),
			CreatedAt:           feed.CreatedAt,
			UpdatedAt:           feed.UpdatedAt,
		})
	}
	return responses, &pagination, nil
}
