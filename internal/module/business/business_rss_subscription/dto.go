// internal/module/business/business_rss_subscription/dto.go
package business_rss_subscription

type CreateUpdateBusinessRSSSubscriptionInput struct {
	Title        string `json:"title" validate:"required"`
	AppRssFeedId int64  `json:"appRssFeedId" validate:"required"`
	IsActive     bool   `json:"isActive" validate:"required"`
}
