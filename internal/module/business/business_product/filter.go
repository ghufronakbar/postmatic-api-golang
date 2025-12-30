package business_product

type GetBusinessProductsByBusinessRootIDFilter struct {
	Search     string  `json:"search"`
	SortBy     string  `json:"sortBy"`
	PageOffset int     `json:"pageOffset"`
	PageLimit  int     `json:"pageLimit"`
	SortDir    string  `json:"sortDir"`
	Page       int     `json:"page"`
	Category   string  `json:"category"`
	DateStart  *string `json:"dateStart"`
	DateEnd    *string `json:"dateEnd"`
}

var SORT_BY = []string{"name", "created_at", "updated_at", "price", "id"}
