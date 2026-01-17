// internal/module/creator/creator_image/filter.go
package creator_image_service

type GetCreatorImageFilter struct {
	Search            string  `json:"search"`
	SortBy            string  `json:"sortBy"`
	PageOffset        int     `json:"pageOffset"`
	PageLimit         int     `json:"pageLimit"`
	SortDir           string  `json:"sortDir"`
	Page              int     `json:"page"`
	DateStart         *string `json:"dateStart"`
	DateEnd           *string `json:"dateEnd"`
	TypeCategoryID    *int64  `json:"typeCategoryId"`
	ProductCategoryID *int64  `json:"productCategoryId"`
	Published         *bool   `json:"published"`
	ProfileID         string  `json:"profileId"`
}

var SORT_BY = []string{"name", "created_at", "updated_at", "id"}
