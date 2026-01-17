// internal/module/affiliator/referral_basic/service.go
package referral_basic_service

import (
	"context"
	crand "crypto/rand"
	"database/sql"

	referral_rule_service "postmatic-api/internal/module/app/referral_rule/service"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/utils"
)

type ReferralBasicService struct {
	store entity.Store
	rule  *referral_rule_service.ReferralService
}

func NewService(store entity.Store, rule *referral_rule_service.ReferralService) *ReferralBasicService {
	return &ReferralBasicService{
		store: store,
		rule:  rule,
	}
}

func (s *ReferralBasicService) GetReferralBasicByProfileId(
	ctx context.Context,
	filter GetReferralBasicFilter,
) (ReferralBasicResponse, error) {

	// 1) fast path
	reff, err := s.store.GetProfileReferralCodeByProfileIdBasic(ctx, filter.ProfileID)
	if err != nil && err != sql.ErrNoRows {
		return ReferralBasicResponse{}, err
	}
	if err == nil {
		return mapReferralToResponse(reff), nil
	}

	// 2) build params
	rule, err := s.rule.GetRuleReferral(ctx)
	if err != nil {
		return ReferralBasicResponse{}, err
	}

	paramsBase := entity.CreateProfileReferralCodeParams{
		ProfileID:         filter.ProfileID,
		Type:              entity.ReferralTypeBasic,
		IsActive:          true,
		TotalDiscount:     rule.TotalDiscount,
		DiscountType:      entity.DiscountType(rule.DiscountType),
		ExpiredDays:       utils.NullInt32ToNullInt32(rule.ExpiredDays),
		MaxDiscount:       rule.MaxDiscount,
		MaxUsage:          utils.NullInt32ToNullInt32(rule.MaxUsage),
		RewardPerReferral: rule.RewardPerReferral,
	}

	// 3) retry loop
	for i := 0; i < 10; i++ {
		code, err := generateReffBasicCode(8)
		if err != nil {
			return ReferralBasicResponse{}, err
		}
		params := paramsBase
		params.Code = code

		created, err := s.store.CreateProfileReferralCode(ctx, params)
		if err == nil {
			return mapReferralToResponse(created), nil
		}

		// cek unique violation (Postgres)
		if utils.IsUniqueViolation(err) {
			// kalau profile sudah keburu punya basic, ambil dan return
			existing, e := s.store.GetProfileReferralCodeByProfileIdBasic(ctx, filter.ProfileID)
			if e == nil {
				return mapReferralToResponse(existing), nil
			}
			if e == sql.ErrNoRows {
				// berarti collision code -> retry
				continue
			}
			return ReferralBasicResponse{}, e
		}

		return ReferralBasicResponse{}, err
	}

	return ReferralBasicResponse{}, errs.NewBadRequest("FAILED_TO_GENERATE_UNIQUE_REFERRAL_CODE")
}

func mapReferralToResponse(r entity.ProfileReferralCode) ReferralBasicResponse {
	var expDays *int32
	if r.ExpiredDays.Valid {
		expDays = &r.ExpiredDays.Int32
	}
	var maxUsage *int32
	if r.MaxUsage.Valid {
		maxUsage = &r.MaxUsage.Int32
	}
	return ReferralBasicResponse{
		ID:                r.ID,
		Code:              r.Code,
		TotalDiscount:     r.TotalDiscount,
		DiscountType:      string(r.DiscountType),
		ExpiredDays:       expDays,
		MaxDiscount:       r.MaxDiscount,
		MaxUsage:          maxUsage,
		RewardPerReferral: r.RewardPerReferral,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}

func generateReffBasicCode(n int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	if _, err := crand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}
