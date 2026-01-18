package referral_basic_service

import (
	"time"

	"github.com/google/uuid"
)

// ValidateReferralInput is input for validating referral code for payment
type ValidateReferralInput struct {
	Code           string
	ProfileID      uuid.UUID
	BusinessRootID int64
}

// ReferralValidationResponse is the result of referral validation
type ReferralValidationResponse struct {
	Valid             bool      `json:"valid"`
	Message           string    `json:"message"`
	ReferralCodeID    int64     `json:"referralCodeId,omitempty"`
	DiscountType      string    `json:"discountType,omitempty"`      // "fixed" atau "percentage"
	TotalDiscount     int64     `json:"totalDiscount,omitempty"`     // nilai diskon (nominal atau %)
	MaxDiscount       int64     `json:"maxDiscount,omitempty"`       // max cap untuk percentage
	OwnerProfileID    uuid.UUID `json:"ownerProfileId,omitempty"`    // untuk reward calculation
	RewardPerReferral int64     `json:"rewardPerReferral,omitempty"` // reward to owner
}

// ReferralCodeDetailResponse is the full response for a referral code
type ReferralCodeDetailResponse struct {
	ID                int64     `json:"id"`
	Code              string    `json:"code"`
	Type              string    `json:"type"`
	IsActive          bool      `json:"isActive"`
	TotalDiscount     int64     `json:"totalDiscount"`
	DiscountType      string    `json:"discountType"`
	ExpiredDays       *int32    `json:"expiredDays"`
	MaxDiscount       int64     `json:"maxDiscount"`
	MaxUsage          *int32    `json:"maxUsage"`
	RewardPerReferral int64     `json:"rewardPerReferral"`
	OwnerProfileID    uuid.UUID `json:"ownerProfileId"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}
