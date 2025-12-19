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

func (s *BusinessInformationService) GetJoinedBusinessesByProfileID(ctx context.Context, profileId string, filterData GetJoinedBusinessesByProfileIDFilter) ([]GetJoinedBusinessesByProfileIDResponse, *pagination.Pagination, error) {
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
		Search:     filterData.Search,
		SortBy:     string(filterData.SortBy),
		PageOffset: int32(filterData.PageOffset),
		PageLimit:  int32(filterData.PageLimit),
		SortDir:    string(filterData.SortDir),
		DateStart:  utils.NullStringToNullTime(filterData.DateStart),
		DateEnd:    utils.NullStringToNullTime(filterData.DateEnd),
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
			AnsweredAt:  v.MemberAnsweredAt.Time,
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
		Search:    filterData.Search,
	}

	count, err := s.store.CountJoinedBusinessesByProfileID(ctx, filterCount)
	if err != nil {
		return nil, nil, errs.NewInternalServerError(err)
	}

	paginationParams := pagination.PaginationParams{
		Total: int(count),
		Page:  filterData.Page,
		Limit: filterData.PageLimit,
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

	var businessRootId string

	e := s.store.ExecTx(ctx, func(tx *entity.Queries) error {

		businessRoot, err := tx.CreateBusinessRoot(ctx)
		if err != nil {
			return err
		}

		businessRootId = businessRoot.String()

		_, err = tx.CreateBusinessKnowledge(ctx, entity.CreateBusinessKnowledgeParams{
			BusinessRootID:     businessRoot,
			Name:               input.BusinessKnowledge.Name,
			PrimaryLogoUrl:     sql.NullString{String: input.BusinessKnowledge.PrimaryLogoUrl, Valid: input.BusinessKnowledge.PrimaryLogoUrl != ""},
			Category:           input.BusinessKnowledge.Category,
			Description:        sql.NullString{String: input.BusinessKnowledge.Description, Valid: input.BusinessKnowledge.Description != ""},
			UniqueSellingPoint: sql.NullString{String: input.BusinessKnowledge.UniqueSellingPoint, Valid: input.BusinessKnowledge.UniqueSellingPoint != ""},
			WebsiteUrl:         sql.NullString{String: input.BusinessKnowledge.WebsiteUrl, Valid: input.BusinessKnowledge.WebsiteUrl != ""},
			VisionMission:      sql.NullString{String: input.BusinessKnowledge.VisionMission, Valid: input.BusinessKnowledge.VisionMission != ""},
			Location:           sql.NullString{String: input.BusinessKnowledge.Location, Valid: input.BusinessKnowledge.Location != ""},
			ColorTone:          sql.NullString{String: input.BusinessKnowledge.ColorTone, Valid: input.BusinessKnowledge.ColorTone != ""},
		})
		if err != nil {
			return err
		}

		_, err = tx.CreateBusinessProduct(ctx, entity.CreateBusinessProductParams{
			BusinessRootID: businessRoot,
			Name:           input.ProductKnowledge.Name,
			Price:          priceStr,
			Description:    sql.NullString{String: input.ProductKnowledge.Description, Valid: input.ProductKnowledge.Description != ""},
			ImageUrls:      input.ProductKnowledge.ImageUrls,
		})

		if err != nil {
			return err
		}

		_, err = tx.CreateBusinessMember(ctx, entity.CreateBusinessMemberParams{
			BusinessRootID: businessRoot,
			ProfileID:      profileUUID,
			Role:           entity.BusinessMemberRoleOwner,
			AnsweredAt:     sql.NullTime{},
			Status:         entity.BusinessMemberStatusPending,
		})

		// TODO: send email notification to owner

		if err != nil {
			return err
		}

		return nil
	})

	if e != nil {
		return SetupBusinessRootFirstTimeResponse{}, errs.NewInternalServerError(e)
	}

	res := SetupBusinessRootFirstTimeResponse{
		ID: businessRootId,
	}

	return res, nil
}

func (s *BusinessInformationService) GetBusinessById(ctx context.Context, businessId string, profileId string) (GetBusinessByIdResponse, error) {
	businessUUID, err := uuid.Parse(businessId)
	if err != nil {
		return GetBusinessByIdResponse{}, errs.NewBadRequest("INVALID_BUSINESS_ID")
	}
	business, err := s.store.GetBusinessKnowledgeByBusinessRootID(ctx, businessUUID)
	fmt.Println(err)
	if err != nil {
		return GetBusinessByIdResponse{}, err
	}

	members, err := s.store.GetMembersByBusinessRootID(ctx, businessUUID)
	if err != nil {
		return GetBusinessByIdResponse{}, err
	}

	var memberBusiness []BusinessMemberSub
	var userProfile *BusinessMemberSub
	for _, m := range members {
		if m.ProfileID.String() == profileId {
			userProfile = &BusinessMemberSub{
				Status: string(m.Status),
				Role:   string(m.Role),
				Profile: ProfileSub{
					ID:       m.ProfileID.String(),
					Name:     m.ProfileName,
					ImageUrl: utils.NullStringToString(m.ProfileImageUrl),
					Email:    m.ProfileEmail,
				},
			}
		}
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

	if userProfile == nil {
		return GetBusinessByIdResponse{}, errs.NewForbidden("USER_NOT_FOUND_IN_BUSINESS")
	}

	res := GetBusinessByIdResponse{
		ID:             business.BusinessRootID.String(),
		Name:           business.Name,
		PrimaryLogoUrl: utils.NullStringToString(business.PrimaryLogoUrl),
		Category:       business.Category,
		Description:    utils.NullStringToString(business.Description),
		ColorTone:      utils.NullStringToString(business.ColorTone),
		CreatedAt:      business.CreatedAt,
		UpdatedAt:      business.UpdatedAt,
		Members:        memberBusiness,
		UserPosition:   *userProfile,
	}
	return res, nil
}
