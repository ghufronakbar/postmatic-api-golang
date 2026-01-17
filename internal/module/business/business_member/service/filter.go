// internal/module/business/business_member/filter.go
package business_member_service

import "github.com/google/uuid"

type GetBusinessMembersByBusinessRootIDFilter struct {
	BusinessRootID int64     `json:"businessRootId"`
	ProfileID      uuid.UUID `json:"profileId"`
	IsVerified     *bool     `json:"isVerified"`
}

var SORT_BY = []string{"name", "created_at", "updated_at", "id"}
