// internal/module/affiliator/referral_basic/viewmodel.go
package referral_basic

import "time"

type ReferralBasicResponse struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	// CONSUMER
	TotalDiscount int64  `json:"totalDiscount"`
	DiscountType  string `json:"discountType"`
	ExpiredDays   *int32 `json:"expiredDays"`
	MaxDiscount   int64  `json:"maxDiscount"`
	MaxUsage      *int32 `json:"maxUsage"`
	// PRODUCER
	RewardPerReferral int64     `json:"rewardPerReferral"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}
