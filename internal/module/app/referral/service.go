// internal/module/app/referral/service.go
package referral

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
		// DEFAULT
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
		var expDays *int32
		if data.ExpiredDays.Valid {
			expDays = &data.ExpiredDays.Int32
		}
		var maxUsage *int32
		if data.MaxUsage.Valid {
			maxUsage = &data.MaxUsage.Int32
		}
		res = RuleReferralResponse{
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
		// IF DEFAULT NO NEED TO CREATE LOGS
		return res, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return res, err
	}

	res = RuleReferralResponse{
		ID:                data.ID,
		TotalDiscount:     data.TotalDiscount,
		DiscountType:      string(data.DiscountType),
		MaxDiscount:       data.MaxDiscount,
		RewardPerReferral: data.RewardPerReferral,
		CreatedAt:         data.CreatedAt,
		UpdatedAt:         data.UpdatedAt,
	}

	if data.ExpiredDays.Valid {
		res.ExpiredDays = &data.ExpiredDays.Int32
	}
	if data.MaxUsage.Valid {
		res.MaxUsage = &data.MaxUsage.Int32
	}
	return res, nil
}

func (s *ReferralService) UpsertRuleReferral(ctx context.Context, input UpsertAppProfileReferralRulesDTO) (RuleReferralResponse, error) {
	var res RuleReferralResponse
	checkProfile, err := s.store.GetProfileById(ctx, input.ProfileID)
	if err == sql.ErrNoRows {
		return res, errs.NewNotFound("PROFILE_NOT_FOUND")
	}
	if err != nil {
		return res, err
	}
	if checkProfile.Role != entity.AppRoleAdmin {
		return res, errs.NewForbidden("")
	}
	allowedDiscountType := map[string]entity.DiscountType{
		"fixed":      entity.DiscountTypeFixed,
		"percentage": entity.DiscountTypePercentage,
	}
	if _, ok := allowedDiscountType[input.DiscountType]; !ok {
		return res, errs.NewBadRequest("DISCOUNT_TYPE_NOT_ALLOWED")
	}
	e := s.store.ExecTx(ctx, func(q *entity.Queries) error {
		data, err := q.UpsertAppProfileReferralRules(ctx, entity.UpsertAppProfileReferralRulesParams{
			TotalDiscount:     input.TotalDiscount,
			DiscountType:      allowedDiscountType[input.DiscountType],
			ExpiredDays:       utils.NullInt32ToNullInt32(input.ExpiredDays),
			MaxDiscount:       input.MaxDiscount,
			MaxUsage:          utils.NullInt32ToNullInt32(input.MaxUsage),
			RewardPerReferral: input.RewardPerReferral,
		})
		if err != nil {
			return err
		}

		_, err = q.InsertAppProfileReferralChange(ctx, entity.InsertAppProfileReferralChangeParams{
			ProfileID:         input.ProfileID,
			TotalDiscount:     input.TotalDiscount,
			DiscountType:      allowedDiscountType[input.DiscountType],
			ExpiredDays:       utils.NullInt32ToNullInt32(input.ExpiredDays),
			MaxDiscount:       input.MaxDiscount,
			MaxUsage:          utils.NullInt32ToNullInt32(input.MaxUsage),
			RewardPerReferral: input.RewardPerReferral,
		})
		if err != nil {
			return err
		}
		res = RuleReferralResponse{
			ID:                data.ID,
			TotalDiscount:     data.TotalDiscount,
			DiscountType:      string(data.DiscountType),
			MaxDiscount:       data.MaxDiscount,
			RewardPerReferral: data.RewardPerReferral,
			CreatedAt:         data.CreatedAt,
			UpdatedAt:         data.UpdatedAt,
		}
		if data.ExpiredDays.Valid {
			res.ExpiredDays = &data.ExpiredDays.Int32
		}
		if data.MaxUsage.Valid {
			res.MaxUsage = &data.MaxUsage.Int32
		}
		return nil
	})
	if e != nil {
		return res, e
	}

	return res, nil
}
