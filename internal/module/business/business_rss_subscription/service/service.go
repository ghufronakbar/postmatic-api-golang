// internal/module/business/business_rss_subscription/service.go
package business_rss_subscription_service

import (
	"context"
	"database/sql"

	rss_service "postmatic-api/internal/module/app/rss/service"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/pagination"
)

type BusinessRssSubscriptionService struct {
	store      entity.Store
	rssService *rss_service.RSSService
}

func NewService(store entity.Store, rssService *rss_service.RSSService) *BusinessRssSubscriptionService {
	return &BusinessRssSubscriptionService{
		store:      store,
		rssService: rssService,
	}
}

func (s *BusinessRssSubscriptionService) GetBusinessRssSubscriptionByBusinessRootID(ctx context.Context, filter GetBusinessRssSubscriptionByBusinessRootIdFilter) ([]BusinessRSSSubscriptionResponse, pagination.Pagination, error) {

	inputFilter := entity.GetBusinessRssSubscriptionsByBusinessRootIDParams{
		BusinessRootID: filter.BusinessRootID,
		Search:         filter.Search,
		SortBy:         string(filter.SortBy),
		PageOffset:     int32(filter.PageOffset),
		PageLimit:      int32(filter.PageLimit),
		SortDir:        string(filter.SortDir),
	}

	rssSubs, err := s.store.GetBusinessRssSubscriptionsByBusinessRootID(ctx, inputFilter)
	if err != nil && err != sql.ErrNoRows {
		return []BusinessRSSSubscriptionResponse{}, pagination.Pagination{}, errs.NewInternalServerError(err)
	}

	if rssSubs == nil {
		return []BusinessRSSSubscriptionResponse{}, pagination.Pagination{}, nil
	}

	var result []BusinessRSSSubscriptionResponse
	for _, v := range rssSubs {
		var catId int64
		if v.CategoryID.Valid {
			catId = v.CategoryID.Int64
		}
		rssCat := AppRssCategory{
			ID:   catId,
			Name: v.CategoryName.String,
		}
		var feedId int64
		if v.FeedID.Valid {
			feedId = v.FeedID.Int64
		}
		rssFeed := AppRssFeedSub{
			ID:             feedId,
			Title:          v.FeedTitle.String,
			AppRssCategory: rssCat,
		}
		result = append(result, BusinessRSSSubscriptionResponse{
			BusinessRootID: filter.BusinessRootID,
			CreatedAt:      v.SubscriptionCreatedAt,
			UpdatedAt:      v.SubscriptionUpdatedAt,
			ID:             v.SubscriptionID,
			Title:          v.SubscriptionTitle,
			IsActive:       v.SubscriptionIsActive,
			AppRssId:       v.SubscriptionAppRssFeedID,
			AppRssFeed:     rssFeed,
		})
	}

	countParam := entity.CountBusinessRssSubscriptionsByBusinessRootIDParams{
		BusinessRootID: filter.BusinessRootID,
		Search:         filter.Search,
	}

	count, err := s.store.CountBusinessRssSubscriptionsByBusinessRootID(ctx, countParam)
	if err != nil {
		return []BusinessRSSSubscriptionResponse{}, pagination.Pagination{}, errs.NewInternalServerError(err)
	}

	paginationParams := pagination.PaginationParams{
		Total: int(count),
		Page:  filter.Page,
		Limit: filter.PageLimit,
	}

	pagination := pagination.NewPagination(&paginationParams)

	return result, pagination, nil
}

func (s *BusinessRssSubscriptionService) CreateBusinessRssSubscription(ctx context.Context, input CreateBusinessRSSSubscriptionInput) (CreateUpdateDeleteResponse, error) {
	appRssFeedId := input.AppRssFeedId

	_, err := s.rssService.GetRSSFeedById(ctx, appRssFeedId)
	if err != nil {
		return CreateUpdateDeleteResponse{}, err
	}

	exist, err := s.store.GetBusinessRssSubscriptionByBusinessRootIdAndAppRssFeedId(ctx,
		entity.GetBusinessRssSubscriptionByBusinessRootIdAndAppRssFeedIdParams{
			BusinessRootID: input.BusinessRootID,
			AppRssFeedID:   appRssFeedId,
		})

	if err != nil && err != sql.ErrNoRows {
		return CreateUpdateDeleteResponse{}, err
	}

	if exist.ID != 0 {
		return CreateUpdateDeleteResponse{}, errs.NewBadRequest("SUBSCRIPTION_ALREADY_EXIST")
	}

	inputParam := entity.CreateBusinessRssSubscriptionParams{
		BusinessRootID: input.BusinessRootID,
		Title:          input.Title,
		IsActive:       input.IsActive,
		AppRssFeedID:   appRssFeedId,
	}

	created, err := s.store.CreateBusinessRssSubscription(ctx, inputParam)
	if err != nil {
		return CreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	return CreateUpdateDeleteResponse{
		ID: created.ID,
	}, nil
}

func (s *BusinessRssSubscriptionService) UpdateBusinessRssSubscription(
	ctx context.Context,
	input UpdateBusinessRSSSubscriptionInput,
) (CreateUpdateDeleteResponse, error) {

	// Validasi feed exists
	if _, err := s.rssService.GetRSSFeedById(ctx, input.AppRssFeedId); err != nil {
		return CreateUpdateDeleteResponse{}, err
	}

	// ✅ Ambil subscription tapi pastikan milik businessRootId
	checkSubscription, err := s.store.GetBusinessRssSubscriptionByIDAndBusinessRootID(
		ctx,
		entity.GetBusinessRssSubscriptionByIDAndBusinessRootIDParams{
			ID:             input.ID,
			BusinessRootID: input.BusinessRootID,
		},
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return CreateUpdateDeleteResponse{}, errs.NewBadRequest("SUBSCRIPTION_NOT_FOUND")
		}
		return CreateUpdateDeleteResponse{}, err
	}

	// cek apakah feed berubah (perhatikan Valid)
	currentFeed := checkSubscription.AppRssFeedID

	// ✅ Kalau feed berubah, cek apakah feed baru sudah di-subscribe row lain
	if currentFeed != input.AppRssFeedId {
		exist, err := s.store.ExistsBusinessRssSubscriptionByBusinessRootIDAndFeedIDExceptID(
			ctx,
			entity.ExistsBusinessRssSubscriptionByBusinessRootIDAndFeedIDExceptIDParams{
				BusinessRootID: input.BusinessRootID,
				AppRssFeedID:   input.AppRssFeedId,
				ID:             input.ID,
			},
		)
		if err != nil {
			return CreateUpdateDeleteResponse{}, err
		}
		if exist {
			return CreateUpdateDeleteResponse{}, errs.NewBadRequest("SUBSCRIPTION_ALREADY_EXIST")
		}
	}

	// ✅ Update sekali saja (baik feed sama maupun beda)
	_, err = s.store.EditBusinessRssSubscription(ctx, entity.EditBusinessRssSubscriptionParams{
		Title:        input.Title,
		IsActive:     input.IsActive,
		AppRssFeedID: input.AppRssFeedId,
		ID:           input.ID,
	})
	if err != nil {
		return CreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	return CreateUpdateDeleteResponse{ID: input.ID}, nil
}

func (s *BusinessRssSubscriptionService) DeleteBusinessRssSubscription(ctx context.Context, subscriptionId int64) (CreateUpdateDeleteResponse, error) {

	check, err := s.store.GetBusinessRssSubscriptionById(ctx, subscriptionId)
	if err != nil && err != sql.ErrNoRows {
		return CreateUpdateDeleteResponse{}, err
	}

	if check.ID == 0 {
		return CreateUpdateDeleteResponse{}, errs.NewBadRequest("SUBSCRIPTION_NOT_FOUND")
	}

	err = s.store.HardDeleteBusinessRssSubscriptionByID(ctx, subscriptionId)
	if err != nil {
		return CreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	return CreateUpdateDeleteResponse{
		ID: check.ID,
	}, nil
}
