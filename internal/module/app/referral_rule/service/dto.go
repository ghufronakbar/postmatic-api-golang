// internal/module/app/referral_rule/dto.go
package referral_rule_service

import "github.com/google/uuid"

type UpsertAppProfileReferralRulesDTO struct {
	// TRACK
	ProfileID uuid.UUID `json:"profileId" validate:"required"`
	// CONSUMER
	TotalDiscount int64  `json:"totalDiscount" validate:"required"`
	DiscountType  string `json:"discountType" validate:"required,oneof=fixed percentage"`
	ExpiredDays   *int32 `json:"expiredDays" validate:"omitempty,gte=1"`
	MaxDiscount   int64  `json:"maxDiscount" validate:"required"`
	MaxUsage      *int32 `json:"maxUsage" validate:"omitempty,gte=1"`
	// PRODUCER
	RewardPerReferral int64 `json:"rewardPerReferral" validate:"required"`
}
