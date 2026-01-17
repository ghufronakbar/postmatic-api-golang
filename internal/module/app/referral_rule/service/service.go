// internal/module/app/referral_rule/service.go
package referral_rule_service

import (
	"context"
	"database/sql"

	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/utils"
)

type ReferralService struct {
	store entity.Store
}

func NewReferralService(store entity.Store) *ReferralService {
	return &ReferralService{store: store}
}

// RULE
func (s *ReferralService) GetRuleReferral(ctx context.Context) (RuleReferralResponse, error) {
	var res RuleReferralResponse

	data, err := s.store.GetAppProfileReferralRules(ctx)
	if err == sql.ErrNoRows {
		// Lazy init default (kalau belum di-seed via migration)
		data, err = s.store.UpsertAppProfileReferralRules(ctx, entity.UpsertAppProfileReferralRulesParams{
			TotalDiscount:     100,
			DiscountType:      entity.DiscountTypePercentage,
			ExpiredDays:       sql.NullInt32{Valid: false},
			MaxDiscount:       20000,
			MaxUsage:          sql.NullInt32{Valid: false},
			RewardPerReferral: 20000,
		})
		if err != nil {
			return res, err
		}
	} else if err != nil {
		return res, err
	}

	return mapRulesToResponse(data), nil
}

func (s *ReferralService) UpsertRuleReferral(ctx context.Context, input UpsertAppProfileReferralRulesDTO) (RuleReferralResponse, error) {
	var res RuleReferralResponse

	// 1) AuthZ: hanya admin
	profile, err := s.store.GetProfileById(ctx, input.ProfileID)
	if err == sql.ErrNoRows {
		return res, errs.NewNotFound("PROFILE_NOT_FOUND")
	}
	if err != nil {
		return res, err
	}
	if profile.Role != entity.AppRoleAdmin {
		return res, errs.NewForbidden("")
	}

	// 2) Validate + normalize
	allowedDiscountType := map[string]entity.DiscountType{
		"fixed":      entity.DiscountTypeFixed,
		"percentage": entity.DiscountTypePercentage,
	}
	dt, ok := allowedDiscountType[input.DiscountType]
	if !ok {
		return res, errs.NewBadRequest("DISCOUNT_TYPE_NOT_ALLOWED")
	}

	// Normalisasi fixed: max_discount selalu = total_discount
	if input.DiscountType == "fixed" {
		input.MaxDiscount = input.TotalDiscount
	}

	// Validasi angka (sesuaikan kalau kamu memang mau allow 0)
	if input.TotalDiscount < 0 {
		return res, errs.NewBadRequest("TOTAL_DISCOUNT_MUST_BE_POSITIVE")
	}
	if input.MaxDiscount < 0 {
		return res, errs.NewBadRequest("MAX_DISCOUNT_MUST_BE_POSITIVE")
	}
	if input.RewardPerReferral < 0 {
		return res, errs.NewBadRequest("REWARD_PER_REFERRAL_MUST_BE_POSITIVE")
	}
	if input.ExpiredDays != nil && *input.ExpiredDays <= 0 {
		return res, errs.NewBadRequest("EXPIRED_DAYS_MUST_BE_POSITIVE_OR_NULL")
	}
	if input.MaxUsage != nil && *input.MaxUsage <= 0 {
		return res, errs.NewBadRequest("MAX_USAGE_MUST_BE_POSITIVE_OR_NULL")
	}
	if input.DiscountType == "percentage" && input.TotalDiscount > 100 {
		return res, errs.NewBadRequest("TOTAL_DISCOUNT_MUST_BE_LESS_OR_EQUAL_100_PERCENT")
	}

	expiredDays := utils.NullInt32ToNullInt32(input.ExpiredDays)
	maxUsage := utils.NullInt32ToNullInt32(input.MaxUsage)

	// 3) Ensure singleton row exists + ambil current state untuk compare
	//    (pakai GetRuleReferral supaya kalau belum ada, otomatis dibuat)
	current, err := s.GetRuleReferral(ctx)
	if err != nil {
		return res, err
	}

	// 4) Compare (hindari no-op supaya nggak insert audit change yang sama)
	changed :=
		current.TotalDiscount != input.TotalDiscount ||
			current.DiscountType != string(dt) ||
			!equalInt32Ptr(current.ExpiredDays, input.ExpiredDays) ||
			current.MaxDiscount != input.MaxDiscount ||
			!equalInt32Ptr(current.MaxUsage, input.MaxUsage) ||
			current.RewardPerReferral != input.RewardPerReferral

	if !changed {
		// Return current state biar response tidak kosong
		return current, nil
	}

	// 5) Transaction: upsert + insert change
	txErr := s.store.ExecTx(ctx, func(q *entity.Queries) error {
		data, err := q.UpsertAppProfileReferralRules(ctx, entity.UpsertAppProfileReferralRulesParams{
			TotalDiscount:     input.TotalDiscount,
			DiscountType:      dt,
			ExpiredDays:       expiredDays,
			MaxDiscount:       input.MaxDiscount,
			MaxUsage:          maxUsage,
			RewardPerReferral: input.RewardPerReferral,
		})
		if err != nil {
			return err
		}

		// Audit snapshot: pakai hasil dari DB (data) supaya 100% konsisten
		_, err = q.InsertAppProfileReferralChange(ctx, entity.InsertAppProfileReferralChangeParams{
			ProfileID:         input.ProfileID,
			TotalDiscount:     data.TotalDiscount,
			DiscountType:      data.DiscountType,
			ExpiredDays:       data.ExpiredDays,
			MaxDiscount:       data.MaxDiscount,
			MaxUsage:          data.MaxUsage,
			RewardPerReferral: data.RewardPerReferral,
		})
		if err != nil {
			return err
		}

		res = mapRulesToResponse(data)
		return nil
	})
	if txErr != nil {
		return res, txErr
	}

	return res, nil
}

// HELPER

func mapRulesToResponse(data entity.AppProfileReferralRule) RuleReferralResponse {
	var expDays *int32
	if data.ExpiredDays.Valid {
		expDays = &data.ExpiredDays.Int32
	}

	var maxUsage *int32
	if data.MaxUsage.Valid {
		maxUsage = &data.MaxUsage.Int32
	}

	return RuleReferralResponse{
		ID:                data.ID,
		TotalDiscount:     data.TotalDiscount,
		DiscountType:      string(data.DiscountType),
		ExpiredDays:       expDays,
		MaxDiscount:       data.MaxDiscount,
		MaxUsage:          maxUsage,
		RewardPerReferral: data.RewardPerReferral,
		CreatedAt:         data.CreatedAt,
		UpdatedAt:         data.UpdatedAt,
	}
}

func equalInt32Ptr(a, b *int32) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
