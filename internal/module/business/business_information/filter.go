package business_information

import "github.com/google/uuid"

type GetJoinedBusinessesByProfileIDFilter struct {
	Search     string    `json:"search"`
	SortBy     string    `json:"sortBy"`
	PageOffset int       `json:"pageOffset"`
	PageLimit  int       `json:"pageLimit"`
	SortDir    string    `json:"sortDir"`
	Page       int       `json:"page"`
	DateStart  *string   `json:"dateStart"`
	DateEnd    *string   `json:"dateEnd"`
	ProfileID  uuid.UUID `json:"profileId"`
}

var SORT_BY = []string{"name", "created_at", "updated_at", "answered_at", "id"}
