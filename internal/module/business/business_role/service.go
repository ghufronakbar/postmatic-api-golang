// internal/module/business/business_knowledge/service.go
package business_role

import (
	"context"
	"database/sql"

	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"

	"github.com/google/uuid"
)

type BusinessRoleService struct {
	store entity.Store
}

// Update Constructor: Minta Token Maker dari main.go
func NewService(store entity.Store) *BusinessRoleService {
	return &BusinessRoleService{
		store: store,
	}
}

func (s *BusinessRoleService) GetBusinessRoleByBusinessRootID(ctx context.Context, businessRootId string) (BusinessRoleResponse, error) {
	businessRootUUID, err := uuid.Parse(businessRootId)
	if err != nil {
		return BusinessRoleResponse{}, errs.NewInternalServerError(err)
	}

	bk, err := s.store.GetBusinessRoleByBusinessRootID(ctx, businessRootUUID)
	if err != nil && err != sql.ErrNoRows {
		return BusinessRoleResponse{}, errs.NewInternalServerError(err)
	}

	var goals string
	if bk.Goals.Valid {
		goals = bk.Goals.String
	}

	var hashtags []string
	if bk.Hashtags == nil {
		hashtags = []string{}
	} else {
		hashtags = bk.Hashtags
	}

	result := BusinessRoleResponse{
		BusinessRootId:  businessRootUUID.String(),
		AudiencePersona: bk.AudiencePersona,
		CallToAction:    bk.CallToAction,
		Goals:           goals,
		Hashtags:        hashtags,
		TargetAudience:  bk.TargetAudience,
		Tone:            bk.Tone,
		CreatedAt:       bk.CreatedAt,
		UpdatedAt:       bk.UpdatedAt,
	}

	return result, nil
}

func (s *BusinessRoleService) UpsertBusinessRoleByBusinessRootID(ctx context.Context, businessRootId string, input UpsertBusinessRoleInput) (BusinessRoleResponse, error) {
	businessRootUUID, err := uuid.Parse(businessRootId)
	if err != nil {
		return BusinessRoleResponse{}, errs.NewInternalServerError(err)
	}

	bk, err := s.store.UpsertBusinessRoleByBusinessRootID(ctx, entity.UpsertBusinessRoleByBusinessRootIDParams{
		BusinessRootID:  businessRootUUID,
		AudiencePersona: input.AudiencePersona,
		CallToAction:    input.CallToAction,
		Goals:           sql.NullString{String: input.Goals, Valid: input.Goals != ""},
		Hashtags:        input.Hashtags,
		TargetAudience:  input.TargetAudience,
		Tone:            input.Tone,
	})
	if err != nil {
		return BusinessRoleResponse{}, errs.NewInternalServerError(err)
	}

	res := BusinessRoleResponse{
		BusinessRootId:  businessRootUUID.String(),
		AudiencePersona: bk.AudiencePersona,
		CallToAction:    bk.CallToAction,
		Goals:           bk.Goals.String,
		Hashtags:        bk.Hashtags,
		TargetAudience:  bk.TargetAudience,
		Tone:            bk.Tone,
		CreatedAt:       bk.CreatedAt,
		UpdatedAt:       bk.UpdatedAt,
	}

	return res, nil
}
