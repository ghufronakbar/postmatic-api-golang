// internal/repository/redis/owned_business_repository/dto.go
package owned_business_repository

import "postmatic-api/internal/repository/entity"

type RedisOwnedBusinessInput struct {
	// Profile ID
	ProfileID string `json:"profileId"`
	// Business Sub
	BusinessSub []RedisBusinessSub `json:"businessSub"`
}

type RedisBusinessSub struct {
	// Member ID
	MemberID string `json:"memberId"`
	// Business Root ID
	BusinessRootID string `json:"businessRootId"`
	// Role
	Role entity.BusinessMemberRole `json:"role"`
}
