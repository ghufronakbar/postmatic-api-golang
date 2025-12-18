// internal/module/business/business_information/service.go
package business_information

import (
	"context"
	"database/sql"
	"fmt"

	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/pagination"

	"postmatic-api/pkg/utils"

	"github.com/google/uuid"
)

type BusinessInformationService struct {
	store entity.Store
}

// Update Constructor: Minta Token Maker dari main.go
func NewService(store entity.Store) *BusinessInformationService {
	return &BusinessInformationService{
		store: store,
	}
}

func (s *BusinessInformationService) GetJoinedBusinessesByProfileID(ctx context.Context, profileId string, filter GetJoinedBusinessesByProfileIDFilter) ([]GetJoinedBusinessesByProfileIDResponse, *pagination.Pagination, error) {
	profileUUID, err := uuid.Parse(profileId)
	if err != nil {
		return nil, nil, errs.NewInternalServerError(err)
	}

	profile, err := s.store.GetProfileById(ctx, profileUUID)
	if err != nil {
		return nil, nil, errs.NewInternalServerError(err)
	}

	filterQuery := entity.GetJoinedBusinessesByProfileIDParams{
		ProfileID:  profileUUID,
		Search:     filter.Search,
		SortBy:     string(filter.SortBy),
		PageOffset: int32(filter.PageOffset),
		PageLimit:  int32(filter.PageLimit),
		SortDir:    string(filter.SortDir),
	}

	res, err := s.store.GetJoinedBusinessesByProfileID(ctx, filterQuery)
	if err != nil {
		return nil, nil, errs.NewInternalServerError(err)
	}

	var businessRootIds []uuid.UUID
	for _, v := range res {
		businessRootIds = append(businessRootIds, v.BusinessRootID)
	}

	members, err := s.store.GetMembersByBusinessRootIDs(ctx, businessRootIds)
	if err != nil {
		return nil, nil, errs.NewInternalServerError(err)
	}

	var result []GetJoinedBusinessesByProfileIDResponse
	for _, v := range res {
		var memberBusiness []BusinessMemberSub
		for _, m := range members {
			if m.BusinessRootID == v.BusinessRootID {
				memberBusiness = append(memberBusiness, BusinessMemberSub{
					Status: string(m.Status),
					Role:   string(m.Role),
					Profile: ProfileSub{
						ID:       m.ProfileID.String(),
						Name:     m.ProfileName,
						ImageUrl: utils.NullStringToString(m.ProfileImageUrl),
						Email:    m.ProfileEmail,
					},
				})
			}
		}
		result = append(result, GetJoinedBusinessesByProfileIDResponse{
			ID:          v.BusinessRootID.String(),
			Name:        v.BusinessName,
			Description: v.BusinessDescription.String,
			CreatedAt:   v.BusinessRootCreatedAt,
			UpdatedAt:   v.BusinessRootUpdatedAt,
			Members:     memberBusiness,
			UserPosition: BusinessMemberSub{
				Status: string(v.MemberStatus),
				Role:   string(v.MemberRole),
				Profile: ProfileSub{
					ID:       profileId,
					Name:     profile.Name,
					ImageUrl: utils.NullStringToString(profile.ImageUrl),
					Email:    profile.Email,
				},
			},
		})
	}

	filterCount := entity.CountJoinedBusinessesByProfileIDParams{
		ProfileID: profileUUID,
		Search:    filter.Search,
	}

	count, err := s.store.CountJoinedBusinessesByProfileID(ctx, filterCount)
	if err != nil {
		return nil, nil, errs.NewInternalServerError(err)
	}

	paginationParams := pagination.PaginationParams{
		Total: int(count),
		Page:  filter.Page,
		Limit: filter.PageLimit,
	}

	pagination := pagination.NewPagination(&paginationParams)

	return result, &pagination, nil
}

func (s *BusinessInformationService) SetupBusinessRootFirstTime(ctx context.Context, profileId string, input BusinessSetupInput) (SetupBusinessRootFirstTimeResponse, error) {
	profileUUID, err := uuid.Parse(profileId)
	if err != nil {
		return SetupBusinessRootFirstTimeResponse{}, errs.NewInternalServerError(err)
	}

	priceStr := fmt.Sprintf("%d.00", input.ProductKnowledge.Price) // 35000 -> "35000.00"

	inputNewBusiness := entity.SetupBusinessRootFirstTimeParams{
		// PROFILE / MEMBER
		ProfileID: profileUUID,

		// BUSINESS KNOWLEDGE
		Name:               input.BusinessKnowledge.Name,
		PrimaryLogoUrl:     sql.NullString{String: input.BusinessKnowledge.PrimaryLogoUrl, Valid: input.BusinessKnowledge.PrimaryLogoUrl != ""},
		Category:           input.BusinessKnowledge.Category,
		Description:        sql.NullString{String: input.BusinessKnowledge.Description, Valid: input.BusinessKnowledge.Description != ""},
		UniqueSellingPoint: sql.NullString{String: input.BusinessKnowledge.UniqueSellingPoint, Valid: input.BusinessKnowledge.UniqueSellingPoint != ""},
		WebsiteUrl:         sql.NullString{String: input.BusinessKnowledge.WebsiteUrl, Valid: input.BusinessKnowledge.WebsiteUrl != ""},
		VisionMission:      sql.NullString{String: input.BusinessKnowledge.VisionMission, Valid: input.BusinessKnowledge.VisionMission != ""},
		Location:           sql.NullString{String: input.BusinessKnowledge.Location, Valid: input.BusinessKnowledge.Location != ""},
		ColorTone:          sql.NullString{String: input.BusinessKnowledge.ColorTone, Valid: input.BusinessKnowledge.ColorTone != ""},

		// ROLE
		TargetAudience:  input.RoleKnowledge.TargetAudience,
		Tone:            input.RoleKnowledge.Tone,
		AudiencePersona: input.RoleKnowledge.AudiencePersona,
		CallToAction:    input.RoleKnowledge.CallToAction,
		Goals:           sql.NullString{String: input.RoleKnowledge.Goals, Valid: input.RoleKnowledge.Goals != ""},
		Column20:        input.RoleKnowledge.Hashtags, // hashtags

		// PRODUCT
		Name_2:        input.ProductKnowledge.Name,
		Category_2:    input.ProductKnowledge.Category,
		Description_2: sql.NullString{String: input.ProductKnowledge.Description, Valid: input.ProductKnowledge.Description != ""},
		Currency:      input.ProductKnowledge.Currency,
		Price:         priceStr,
		Column16:      input.ProductKnowledge.ImageUrls, // image urls

	}

	newBusiness, err := s.store.SetupBusinessRootFirstTime(ctx, inputNewBusiness)
	if err != nil {
		return SetupBusinessRootFirstTimeResponse{}, errs.NewInternalServerError(err)
	}

	res := SetupBusinessRootFirstTimeResponse{
		ID: newBusiness.BusinessRootID.String(),
	}

	return res, nil
}
