package rss

type GetRSSFeedFilter struct {
	Search     string `json:"search"`
	SortBy     string `json:"sortBy"`
	PageOffset int    `json:"pageOffset"`
	PageLimit  int    `json:"pageLimit"`
	SortDir    string `json:"sortDir"`
	Page       int    `json:"page"`
	Category   int64  `json:"category"`
}

var SORT_BY_RSS_FEED = []string{"title", "created_at", "updated_at"}

type GetRSSCategoryFilter struct {
	Search     string `json:"search"`
	SortBy     string `json:"sortBy"`
	PageOffset int    `json:"pageOffset"`
	PageLimit  int    `json:"pageLimit"`
	SortDir    string `json:"sortDir"`
	Page       int    `json:"page"`
}

var SORT_BY_RSS_CATEGORY = []string{"name", "created_at", "updated_at", "id"}
