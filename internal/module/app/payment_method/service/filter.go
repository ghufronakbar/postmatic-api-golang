package payment_method_service

type GetPaymentMethodsFilter struct {
	Search     string `json:"search"`
	SortBy     string `json:"sortBy"`
	SortDir    string `json:"sortDir"`
	PageOffset int    `json:"pageOffset"`
	PageLimit  int    `json:"pageLimit"`
	Page       int    `json:"page"`
	IsAdmin    bool   `json:"isAdmin"`
}

var SORT_BY = []string{"name", "code", "created_at", "updated_at", "id"}
