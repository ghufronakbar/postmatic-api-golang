// internal/module/app/rss/service.go
package rss

import (
	"context"
	"database/sql"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/pagination"
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
			ID:        category.ID,
			Name:      category.Name,
			CreatedAt: category.CreatedAt,
			UpdatedAt: category.UpdatedAt,
		})
	}
	return responses, &pagination, nil
}

func (s *RSSService) GetRSSFeed(ctx context.Context, filter GetRSSFeedFilter) ([]RSSResponse, *pagination.Pagination, error) {

	filterData := entity.GetAllRSSFeedParams{
		Search:     filter.Search,
		SortBy:     filter.SortBy,
		SortDir:    filter.SortDir,
		PageOffset: int32(filter.PageOffset),
		PageLimit:  int32(filter.PageLimit),
		Category:   sql.NullInt64{Int64: filter.Category, Valid: filter.Category != 0},
	}
	feeds, err := s.store.GetAllRSSFeed(ctx, filterData)
	if err != nil {
		return nil, nil, err
	}
	filterCount := entity.CountAllRSSFeedParams{
		Search:   filter.Search,
		Category: sql.NullInt64{Int64: filter.Category, Valid: filter.Category != 0},
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
			ID:                  feed.ID,
			Title:               feed.Title,
			URL:                 feed.Url,
			Publisher:           feed.Publisher,
			MasterRSSCategoryID: feed.AppRssCategoryID,
			CreatedAt:           feed.CreatedAt,
			UpdatedAt:           feed.UpdatedAt,
		})
	}
	return responses, &pagination, nil
}

func (s *RSSService) GetRSSFeedById(ctx context.Context, id int64) (RSSResponse, error) {
	feed, err := s.store.GetRssFeedById(ctx, id)
	if err == sql.ErrNoRows {
		return RSSResponse{}, errs.NewNotFound("RSS_FEED_NOT_FOUND")
	}
	if err != nil && err != sql.ErrNoRows {
		return RSSResponse{}, err
	}
	return RSSResponse{
		ID:                  feed.ID,
		Title:               feed.Title,
		URL:                 feed.Url,
		Publisher:           feed.Publisher,
		MasterRSSCategoryID: feed.AppRssCategoryID,
		CreatedAt:           feed.CreatedAt,
		UpdatedAt:           feed.UpdatedAt,
	}, nil
}
