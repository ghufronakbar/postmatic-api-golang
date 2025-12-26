// internal/module/app/category_creator_image/filter.go
package category_creator_image

type GetCategoryCreatorImageTypeFilter struct {
	Search     string `json:"search"`
	SortBy     string `json:"sortBy"`
	PageOffset int    `json:"pageOffset"`
	PageLimit  int    `json:"pageLimit"`
	SortDir    string `json:"sortDir"`
	Page       int    `json:"page"`
}

var SORT_BY_CATEGORY_CREATOR_IMAGE_TYPE = []string{"name", "created_at", "updated_at"}

type GetCategoryCreatorImageProductFilter struct {
	Search     string `json:"search"`
	SortBy     string `json:"sortBy"`
	PageOffset int    `json:"pageOffset"`
	PageLimit  int    `json:"pageLimit"`
	SortDir    string `json:"sortDir"`
	Page       int    `json:"page"`
	Locale     string `json:"locale"` // "id" or "en"
}

var SORT_BY_CATEGORY_CREATOR_IMAGE_PRODUCT = []string{"name", "created_at", "updated_at"}
