// internal/module/business/business_rss_subscription/viewmodel.go
package business_rss_subscription_service

import "time"

type BusinessRSSSubscriptionResponse struct {
	ID             int64         `json:"id"`
	BusinessRootID int64         `json:"businessRootId"`
	Title          string        `json:"title" validate:"required"`
	AppRssId       int64         `json:"appRssId" validate:"required"`
	IsActive       bool          `json:"isActive" validate:"required"`
	AppRssFeed     AppRssFeedSub `json:"appRssFeed"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

type AppRssFeedSub struct {
	ID             int64          `json:"id" validate:"required"`
	Title          string         `json:"title" validate:"required"`
	AppRssCategory AppRssCategory `json:"appRssCategory"`
}

type AppRssCategory struct {
	ID   int64  `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type CreateUpdateDeleteResponse struct {
	ID int64 `json:"id" validate:"required"`
}
