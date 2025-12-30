// internal/module/business/business_knowledge/service.go
package business_knowledge

import (
	"context"
	"database/sql"

	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"

	"postmatic-api/pkg/utils"
)

type BusinessKnowledgeService struct {
	store entity.Store
}

// Update Constructor: Minta Token Maker dari main.go
func NewService(store entity.Store) *BusinessKnowledgeService {
	return &BusinessKnowledgeService{
		store: store,
	}
}

func (s *BusinessKnowledgeService) GetBusinessKnowledgeByBusinessRootID(ctx context.Context, businessRootId int64) (BusinessKnowledgeResponse, error) {

	bk, err := s.store.GetBusinessKnowledgeByBusinessRootID(ctx, businessRootId)
	if err != nil && err != sql.ErrNoRows {
		return BusinessKnowledgeResponse{}, errs.NewInternalServerError(err)
	}

	var websiteUrl *string
	if bk.WebsiteUrl.Valid {
		websiteUrl = &bk.WebsiteUrl.String
	}

	result := BusinessKnowledgeResponse{
		RootBusinessId:     businessRootId,
		Name:               bk.Name,
		PrimaryLogoUrl:     bk.PrimaryLogoUrl.String,
		Description:        bk.Description.String,
		CreatedAt:          bk.CreatedAt,
		UpdatedAt:          bk.UpdatedAt,
		Category:           bk.Category,
		ColorTone:          bk.ColorTone.String,
		UniqueSellingPoint: bk.UniqueSellingPoint.String,
		WebsiteUrl:         websiteUrl,
		VisionMission:      bk.VisionMission.String,
		Location:           bk.Location.String,
	}

	return result, nil
}

func (s *BusinessKnowledgeService) UpsertBusinessKnowledgeByBusinessRootID(ctx context.Context, businessRootId int64, input UpsertBusinessKnowledgeInput) (BusinessKnowledgeResponse, error) {
	var websiteUrl string
	if input.WebsiteUrl != nil {
		websiteUrl = *input.WebsiteUrl
	}

	bk, err := s.store.UpsertBusinessKnowledgeByBusinessRootID(ctx, entity.UpsertBusinessKnowledgeByBusinessRootIDParams{
		BusinessRootID:     businessRootId,
		Name:               input.Name,
		PrimaryLogoUrl:     sql.NullString{String: input.PrimaryLogoUrl, Valid: input.PrimaryLogoUrl != ""},
		Category:           input.Category,
		Description:        sql.NullString{String: input.Description, Valid: input.Description != ""},
		UniqueSellingPoint: sql.NullString{String: input.UniqueSellingPoint, Valid: input.UniqueSellingPoint != ""},
		WebsiteUrl:         sql.NullString{String: websiteUrl, Valid: websiteUrl != ""},
		VisionMission:      sql.NullString{String: input.VisionMission, Valid: input.VisionMission != ""},
		Location:           sql.NullString{String: input.Location, Valid: input.Location != ""},
		ColorTone:          sql.NullString{String: input.ColorTone, Valid: input.ColorTone != ""},
	})
	if err != nil {
		return BusinessKnowledgeResponse{}, errs.NewInternalServerError(err)
	}

	res := BusinessKnowledgeResponse{
		RootBusinessId:     businessRootId,
		Name:               bk.Name,
		PrimaryLogoUrl:     bk.PrimaryLogoUrl.String,
		Description:        bk.Description.String,
		CreatedAt:          bk.CreatedAt,
		UpdatedAt:          bk.UpdatedAt,
		Category:           bk.Category,
		ColorTone:          bk.ColorTone.String,
		UniqueSellingPoint: bk.UniqueSellingPoint.String,
		WebsiteUrl:         utils.NullStringToString(bk.WebsiteUrl),
		VisionMission:      bk.VisionMission.String,
		Location:           bk.Location.String,
	}

	return res, nil
}
