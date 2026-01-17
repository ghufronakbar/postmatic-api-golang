// internal/module/business/business_rss_subscription/filter.go
package business_rss_subscription_service

type GetBusinessRssSubscriptionByBusinessRootIdFilter struct {
	Search         string `json:"search"`
	SortBy         string `json:"sortBy"`
	PageOffset     int    `json:"pageOffset"`
	PageLimit      int    `json:"pageLimit"`
	SortDir        string `json:"sortDir"`
	Page           int    `json:"page"`
	BusinessRootID int64  `json:"businessRootId"`
}

var SORT_BY = []string{"title", "created_at", "updated_at", "id"}
