// internal/repository/redis/owned_business_repository/viewmodel.go
package owned_business_repository

import "postmatic-api/internal/repository/entity"

type RedisOwnedBusinessResponse struct {
	// Member ID
	MemberID int64 `json:"memberId"`
	// Business Root ID
	BusinessRootID int64 `json:"businessRootId"`
	// Role
	Role entity.BusinessMemberRole `json:"role"`
}
