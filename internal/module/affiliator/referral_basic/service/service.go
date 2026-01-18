// internal/module/affiliator/referral_basic/service.go
package referral_basic_service

import (
	"context"
	crand "crypto/rand"
	"database/sql"
	"time"

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

// GetReferralCodeByCode returns the referral code detail by code
func (s *ReferralBasicService) GetReferralCodeByCode(ctx context.Context, code string) (ReferralCodeDetailResponse, error) {
	reff, err := s.store.GetProfileReferralCodeByCode(ctx, code)
	if err == sql.ErrNoRows {
		return ReferralCodeDetailResponse{}, errs.NewNotFound("REFERRAL_CODE_NOT_FOUND")
	}
	if err != nil {
		return ReferralCodeDetailResponse{}, errs.NewInternalServerError(err)
	}

	return mapReferralCodeToDetail(reff), nil
}

// ValidateReferralForPayment validates if a referral code can be used for payment
func (s *ReferralBasicService) ValidateReferralForPayment(ctx context.Context, input ValidateReferralInput) (*ReferralValidationResponse, error) {
	// 1. Get referral code
	reff, err := s.store.GetProfileReferralCodeByCode(ctx, input.Code)
	if err == sql.ErrNoRows {
		return &ReferralValidationResponse{Valid: false, Message: "REFERRAL_CODE_NOT_FOUND"}, nil
	}
	if err != nil {
		return nil, errs.NewInternalServerError(err)
	}

	// 2. Check if active
	if !reff.IsActive {
		return &ReferralValidationResponse{Valid: false, Message: "REFERRAL_CODE_INACTIVE"}, nil
	}

	// 3. Check self-referral (cannot use own code)
	if reff.ProfileID == input.ProfileID {
		return &ReferralValidationResponse{Valid: false, Message: "CANNOT_USE_OWN_REFERRAL_CODE"}, nil
	}

	// 4. Check if expired
	if reff.ExpiredDays.Valid {
		expiryDate := reff.CreatedAt.AddDate(0, 0, int(reff.ExpiredDays.Int32))
		if time.Now().After(expiryDate) {
			return &ReferralValidationResponse{Valid: false, Message: "REFERRAL_CODE_EXPIRED"}, nil
		}
	}

	// 5. Check max usage limit
	if reff.MaxUsage.Valid {
		usageCount, err := s.store.CountReferralCodeUsage(ctx, reff.ID)
		if err != nil {
			return nil, errs.NewInternalServerError(err)
		}
		if int(usageCount) >= int(reff.MaxUsage.Int32) {
			return &ReferralValidationResponse{Valid: false, Message: "REFERRAL_CODE_MAX_USAGE_REACHED"}, nil
		}
	}

	// 6. Check if profile already used this code
	profileUsed, err := s.store.CheckProfileUsedReferralCode(ctx, entity.CheckProfileUsedReferralCodeParams{
		ConsumerProfileID:     input.ProfileID,
		ProfileReferralCodeID: reff.ID,
	})
	if err != nil {
		return nil, errs.NewInternalServerError(err)
	}
	if profileUsed {
		return &ReferralValidationResponse{Valid: false, Message: "PROFILE_ALREADY_USED_REFERRAL_CODE"}, nil
	}

	// 7. Check if business already used this code
	businessUsed, err := s.store.CheckBusinessUsedReferralCode(ctx, entity.CheckBusinessUsedReferralCodeParams{
		BusinessRootID:        input.BusinessRootID,
		ProfileReferralCodeID: reff.ID,
	})
	if err != nil {
		return nil, errs.NewInternalServerError(err)
	}
	if businessUsed {
		return &ReferralValidationResponse{Valid: false, Message: "BUSINESS_ALREADY_USED_REFERRAL_CODE"}, nil
	}

	// All validations passed
	return &ReferralValidationResponse{
		Valid:             true,
		Message:           "REFERRAL_CODE_VALID",
		ReferralCodeID:    reff.ID,
		DiscountType:      string(reff.DiscountType),
		TotalDiscount:     reff.TotalDiscount,
		MaxDiscount:       reff.MaxDiscount,
		OwnerProfileID:    reff.ProfileID,
		RewardPerReferral: reff.RewardPerReferral,
	}, nil
}

func mapReferralCodeToDetail(r entity.ProfileReferralCode) ReferralCodeDetailResponse {
	var expDays *int32
	if r.ExpiredDays.Valid {
		expDays = &r.ExpiredDays.Int32
	}
	var maxUsage *int32
	if r.MaxUsage.Valid {
		maxUsage = &r.MaxUsage.Int32
	}
	return ReferralCodeDetailResponse{
		ID:                r.ID,
		Code:              r.Code,
		Type:              string(r.Type),
		IsActive:          r.IsActive,
		TotalDiscount:     r.TotalDiscount,
		DiscountType:      string(r.DiscountType),
		ExpiredDays:       expDays,
		MaxDiscount:       r.MaxDiscount,
		MaxUsage:          maxUsage,
		RewardPerReferral: r.RewardPerReferral,
		OwnerProfileID:    r.ProfileID,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}
