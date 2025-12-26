// internal/module/business/business_timezone_pref/service.go
package business_timezone_pref

import (
	"context"
	"database/sql"
	"time"

	"postmatic-api/internal/module/app/timezone"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"

	"github.com/google/uuid"
)

type BusinessTimezonePrefService struct {
	store    entity.Store
	timezone *timezone.TimezoneService
}

func NewService(store entity.Store, timezone *timezone.TimezoneService) *BusinessTimezonePrefService {
	return &BusinessTimezonePrefService{
		store:    store,
		timezone: timezone,
	}
}

func (s *BusinessTimezonePrefService) GetBusinessTimezonePrefByBusinessRootID(ctx context.Context, businessRootId string) (BusinessTimezonePrefResponse, error) {
	businessRootUUID, err := uuid.Parse(businessRootId)
	if err != nil {
		return BusinessTimezonePrefResponse{}, errs.NewInternalServerError(err)
	}

	tzDb, err := s.store.GetBusinessTimezonePrefByBusinessRootId(ctx, businessRootUUID)
	if err != nil && err != sql.ErrNoRows {
		return BusinessTimezonePrefResponse{}, errs.NewInternalServerError(err)
	}

	if err == sql.ErrNoRows {
		tzDb = entity.BusinessTimezonePref{
			Timezone:       "Asia/Jakarta",
			ID:             tzDb.ID,
			BusinessRootID: businessRootUUID,
			CreatedAt:      sql.NullTime{Time: time.Now(), Valid: true},
			UpdatedAt:      sql.NullTime{Time: time.Now(), Valid: true},
		}
	}

	tz, err := s.timezone.ValidateTimezone(ctx, tzDb.Timezone)
	if err != nil {
		return BusinessTimezonePrefResponse{}, err
	}

	return BusinessTimezonePrefResponse{
		RootBusinessId: tzDb.BusinessRootID.String(),
		Timezone:       tz.Name,
		Offset:         tz.Offset,
		Label:          tz.Label,
	}, nil
}

func (s *BusinessTimezonePrefService) UpsertBusinessTimezonePrefByBusinessRootID(ctx context.Context, businessRootId string, input UpsertBusinessTimezonePrefInput) (BusinessTimezonePrefResponse, error) {
	businessRootUUID, err := uuid.Parse(businessRootId)
	if err != nil {
		return BusinessTimezonePrefResponse{}, errs.NewInternalServerError(err)
	}

	tzDb, err := s.store.UpsertBusinessTimezonePref(ctx, entity.UpsertBusinessTimezonePrefParams{
		BusinessRootID: businessRootUUID,
		Timezone:       input.Timezone,
	})
	if err != nil {
		return BusinessTimezonePrefResponse{}, errs.NewInternalServerError(err)
	}

	tz, err := s.timezone.ValidateTimezone(ctx, tzDb.Timezone)
	if err != nil {
		return BusinessTimezonePrefResponse{}, err
	}

	return BusinessTimezonePrefResponse{
		RootBusinessId: tzDb.BusinessRootID.String(),
		Timezone:       tz.Name,
		Offset:         tz.Offset,
		Label:          tz.Label,
	}, nil
}
