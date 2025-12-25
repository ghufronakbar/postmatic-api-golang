// internal/module/business/business_rss_subscription/viewmodel.go
package business_rss_subscription

import "time"

type BusinessRSSSubscriptionResponse struct {
	ID             string        `json:"id"`
	BusinessRootID string        `json:"businessRootId"`
	Title          string        `json:"title" validate:"required"`
	AppRssId       string        `json:"appRssId" validate:"required"`
	IsActive       bool          `json:"isActive" validate:"required"`
	AppRssFeed     AppRssFeedSub `json:"appRssFeed"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

type AppRssFeedSub struct {
	ID             string         `json:"id" validate:"required"`
	Title          string         `json:"title" validate:"required"`
	AppRssCategory AppRssCategory `json:"appRssCategory"`
}

type AppRssCategory struct {
	ID   string `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type CreateUpdateDeleteResponse struct {
	ID string `json:"id" validate:"required"`
}
