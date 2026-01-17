// internal/module/business/business_image_content/filter.go
package business_image_content_service

type GetBusinessImageContentsByBusinessRootIDFilter struct {
	BusinessRootID int64
	Search         string  `json:"search"`
	SortBy         string  `json:"sortBy"`
	PageOffset     int     `json:"pageOffset"`
	PageLimit      int     `json:"pageLimit"`
	SortDir        string  `json:"sortDir"`
	Page           int     `json:"page"`
	ReadyToPost    *bool   `json:"readyToPost"` // readyToPosts or not or neither
	DateStart      *string `json:"dateStart"`
	DateEnd        *string `json:"dateEnd"`
}

var SORT_BY = []string{"caption", "created_at", "updated_at", "id"}
