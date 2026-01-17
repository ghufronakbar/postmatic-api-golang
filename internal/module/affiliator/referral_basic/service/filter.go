// internal/module/affiliator/referral_basic/filter.go
package referral_basic_service

import "github.com/google/uuid"

type GetReferralBasicFilter struct {
	ProfileID uuid.UUID `json:"profileId"`
}
