// internal/module/business/business_rss_subscription/service.go
package business_rss_subscription

import (
	"context"
	"database/sql"

	"postmatic-api/internal/module/app/rss"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/pagination"

	"github.com/google/uuid"
)

type BusinessRssSubscriptionService struct {
	store      entity.Store
	rssService *rss.RSSService
}

func NewService(store entity.Store, rssService *rss.RSSService) *BusinessRssSubscriptionService {
	return &BusinessRssSubscriptionService{
		store:      store,
		rssService: rssService,
	}
}

func (s *BusinessRssSubscriptionService) GetBusinessRssSubscriptionByBusinessRootID(ctx context.Context, businessRootId string, filter GetBusinessRssSubscriptionByBusinessRootIdFilter) ([]BusinessRSSSubscriptionResponse, pagination.Pagination, error) {
	businessRootUUID, err := uuid.Parse(businessRootId)
	if err != nil {
		return []BusinessRSSSubscriptionResponse{}, pagination.Pagination{}, errs.NewInternalServerError(err)
	}

	inputFilter := entity.GetBusinessRssSubscriptionsByBusinessRootIDParams{
		BusinessRootID: businessRootUUID,
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
		rssCat := AppRssCategory{
			ID:   v.CategoryID.UUID.String(),
			Name: v.CategoryName.String,
		}
		rssFeed := AppRssFeedSub{
			ID:             v.FeedID.UUID.String(),
			Title:          v.FeedTitle.String,
			AppRssCategory: rssCat,
		}
		result = append(result, BusinessRSSSubscriptionResponse{
			BusinessRootID: businessRootUUID.String(),
			CreatedAt:      v.SubscriptionCreatedAt,
			UpdatedAt:      v.SubscriptionUpdatedAt,
			ID:             v.SubscriptionID.String(),
			Title:          v.SubscriptionTitle,
			IsActive:       v.SubscriptionIsActive,
			AppRssId:       v.SubscriptionAppRssFeedID.UUID.String(),
			AppRssFeed:     rssFeed,
		})
	}

	countParam := entity.CountBusinessRssSubscriptionsByBusinessRootIDParams{
		BusinessRootID: businessRootUUID,
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

func (s *BusinessRssSubscriptionService) CreateBusinessRssSubscription(ctx context.Context, businessRootId string, input CreateUpdateBusinessRSSSubscriptionInput) (CreateUpdateDeleteResponse, error) {
	businessRootUUID, err := uuid.Parse(businessRootId)
	if err != nil {
		return CreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}
	appRssFeedId, err := uuid.Parse(input.AppRssFeedId)
	if err != nil {
		return CreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	_, err = s.rssService.GetRSSFeedById(ctx, input.AppRssFeedId)
	if err != nil {
		return CreateUpdateDeleteResponse{}, err
	}

	exist, err := s.store.GetBusinessRssSubscriptionByBusinessRootIdAndAppRssFeedId(ctx,
		entity.GetBusinessRssSubscriptionByBusinessRootIdAndAppRssFeedIdParams{
			BusinessRootID: businessRootUUID,
			AppRssFeedID:   uuid.NullUUID{UUID: appRssFeedId, Valid: appRssFeedId != uuid.Nil},
		})

	if err != nil && err != sql.ErrNoRows {
		return CreateUpdateDeleteResponse{}, err
	}

	if exist.ID != uuid.Nil {
		return CreateUpdateDeleteResponse{}, errs.NewBadRequest("SUBSCRIPTION_ALREADY_EXIST")
	}

	inputParam := entity.CreateBusinessRssSubscriptionParams{
		BusinessRootID: businessRootUUID,
		Title:          input.Title,
		IsActive:       input.IsActive,
		AppRssFeedID:   uuid.NullUUID{UUID: appRssFeedId, Valid: appRssFeedId != uuid.Nil},
	}

	created, err := s.store.CreateBusinessRssSubscription(ctx, inputParam)
	if err != nil {
		return CreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	return CreateUpdateDeleteResponse{
		ID: created.ID.String(),
	}, nil
}

func (s *BusinessRssSubscriptionService) UpdateBusinessRssSubscription(
	ctx context.Context,
	businessRootId, subscriptionId string,
	input CreateUpdateBusinessRSSSubscriptionInput,
) (CreateUpdateDeleteResponse, error) {

	businessRootUUID, err := uuid.Parse(businessRootId)
	if err != nil {
		return CreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	subscriptionUUID, err := uuid.Parse(subscriptionId)
	if err != nil {
		return CreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	appRssFeedUUID, err := uuid.Parse(input.AppRssFeedId)
	if err != nil {
		return CreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	// Validasi feed exists
	if _, err := s.rssService.GetRSSFeedById(ctx, input.AppRssFeedId); err != nil {
		return CreateUpdateDeleteResponse{}, err
	}

	// ✅ Ambil subscription tapi pastikan milik businessRootId
	checkSubscription, err := s.store.GetBusinessRssSubscriptionByIDAndBusinessRootID(
		ctx,
		entity.GetBusinessRssSubscriptionByIDAndBusinessRootIDParams{
			ID:             subscriptionUUID,
			BusinessRootID: businessRootUUID,
		},
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return CreateUpdateDeleteResponse{}, errs.NewBadRequest("SUBSCRIPTION_NOT_FOUND")
		}
		return CreateUpdateDeleteResponse{}, err
	}

	// cek apakah feed berubah (perhatikan Valid)
	currentFeed := uuid.Nil
	if checkSubscription.AppRssFeedID.Valid {
		currentFeed = checkSubscription.AppRssFeedID.UUID
	}

	// ✅ Kalau feed berubah, cek apakah feed baru sudah di-subscribe row lain
	if currentFeed != appRssFeedUUID {
		exist, err := s.store.ExistsBusinessRssSubscriptionByBusinessRootIDAndFeedIDExceptID(
			ctx,
			entity.ExistsBusinessRssSubscriptionByBusinessRootIDAndFeedIDExceptIDParams{
				BusinessRootID: businessRootUUID,
				AppRssFeedID:   uuid.NullUUID{UUID: appRssFeedUUID, Valid: true},
				ID:             subscriptionUUID,
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
		AppRssFeedID: uuid.NullUUID{UUID: appRssFeedUUID, Valid: true},
		ID:           subscriptionUUID,
	})
	if err != nil {
		return CreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	return CreateUpdateDeleteResponse{ID: subscriptionUUID.String()}, nil
}

func (s *BusinessRssSubscriptionService) DeleteBusinessRssSubscription(ctx context.Context, subscriptionId string) (CreateUpdateDeleteResponse, error) {

	subscriptionUUID, err := uuid.Parse(subscriptionId)
	if err != nil {
		return CreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	check, err := s.store.GetBusinessRssSubscriptionById(ctx, subscriptionUUID)
	if err != nil && err != sql.ErrNoRows {
		return CreateUpdateDeleteResponse{}, err
	}

	if check.ID == uuid.Nil {
		return CreateUpdateDeleteResponse{}, errs.NewBadRequest("SUBSCRIPTION_NOT_FOUND")
	}

	err = s.store.HardDeleteBusinessRssSubscriptionByID(ctx, subscriptionUUID)
	if err != nil {
		return CreateUpdateDeleteResponse{}, errs.NewInternalServerError(err)
	}

	return CreateUpdateDeleteResponse{
		ID: check.ID.String(),
	}, nil
}
