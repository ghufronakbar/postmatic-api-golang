// internal/module/affiliator/referral_basic/filter.go
package referral_basic

import "github.com/google/uuid"

type GetReferralBasicFilter struct {
	ProfileID uuid.UUID `json:"profileId"`
}
