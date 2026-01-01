// internal/module/business/business_rss_subscription/dto.go
package business_rss_subscription

type CreateBusinessRSSSubscriptionInput struct {
	BusinessRootID int64  `json:"businessRootId" validate:"required"`
	Title          string `json:"title" validate:"required"`
	AppRssFeedId   int64  `json:"appRssFeedId" validate:"required"`
	IsActive       bool   `json:"isActive" validate:"required"`
}
type UpdateBusinessRSSSubscriptionInput struct {
	ID             int64  `json:"id" validate:"required"`
	BusinessRootID int64  `json:"businessRootId" validate:"required"`
	Title          string `json:"title" validate:"required"`
	AppRssFeedId   int64  `json:"appRssFeedId" validate:"required"`
	IsActive       bool   `json:"isActive" validate:"required"`
}
