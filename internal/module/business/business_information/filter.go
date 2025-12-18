package business_information

type GetJoinedBusinessesByProfileIDFilter struct {
	Search     string `json:"search"`
	SortBy     string `json:"sortBy"`
	PageOffset int    `json:"pageOffset"`
	PageLimit  int    `json:"pageLimit"`
	SortDir    string `json:"sortDir"`
	Page       int    `json:"page"`
}

var SORT_BY = []string{"name", "created_at", "updated_at"}
