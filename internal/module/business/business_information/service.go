// internal/module/business/business_information/service.go
package business_information

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"postmatic-api/internal/repository/entity"
	"postmatic-api/internal/repository/redis/owned_business_repository"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/pagination"

	"postmatic-api/pkg/utils"

	"github.com/google/uuid"
)

type BusinessInformationService struct {
	store  entity.Store
	obRepo *owned_business_repository.OwnedBusinessRepository
}

// Update Constructor: Minta Token Maker dari main.go
func NewService(store entity.Store, obRepo *owned_business_repository.OwnedBusinessRepository) *BusinessInformationService {
	return &BusinessInformationService{
		store:  store,
		obRepo: obRepo,
	}
}

func (s *BusinessInformationService) GetJoinedBusinessesByProfileID(ctx context.Context, filterData GetJoinedBusinessesByProfileIDFilter) ([]GetJoinedBusinessesByProfileIDResponse, *pagination.Pagination, error) {

	profile, err := s.store.GetProfileById(ctx, filterData.ProfileID)
	if err != nil {
		return nil, nil, errs.NewInternalServerError(err)
	}

	filterQuery := entity.GetJoinedBusinessesByProfileIDParams{
		ProfileID:  profile.ID,
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

	var businessRootIds []int64
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
			ID:          v.BusinessRootID,
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
					ID:       filterData.ProfileID.String(),
					Name:     profile.Name,
					ImageUrl: utils.NullStringToString(profile.ImageUrl),
					Email:    profile.Email,
				},
			},
		})
	}

	filterCount := entity.CountJoinedBusinessesByProfileIDParams{
		ProfileID: filterData.ProfileID,
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

func (s *BusinessInformationService) SetupBusinessRootFirstTime(ctx context.Context, input BusinessSetupInput) (SetupBusinessRootFirstTimeResponse, error) {

	var businessRootId int64
	var memberID int64

	e := s.store.ExecTx(ctx, func(tx *entity.Queries) error {

		businessRootId, err := tx.CreateBusinessRoot(ctx)
		if err != nil {
			return err
		}

		_, err = tx.CreateBusinessKnowledge(ctx, entity.CreateBusinessKnowledgeParams{
			BusinessRootID:     businessRootId,
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
			BusinessRootID: businessRootId,
			Name:           input.ProductKnowledge.Name,
			Price:          input.ProductKnowledge.Price,
			Description:    sql.NullString{String: input.ProductKnowledge.Description, Valid: input.ProductKnowledge.Description != ""},
			ImageUrls:      input.ProductKnowledge.ImageUrls,
			Category:       input.ProductKnowledge.Category,
			Currency:       input.ProductKnowledge.Currency,
		})

		if err != nil {
			return err
		}

		_, err = tx.CreateBusinessRole(ctx, entity.CreateBusinessRoleParams{
			BusinessRootID:  businessRootId,
			TargetAudience:  input.RoleKnowledge.TargetAudience,
			Tone:            input.RoleKnowledge.Tone,
			AudiencePersona: input.RoleKnowledge.AudiencePersona,
			Hashtags:        input.RoleKnowledge.Hashtags,
			CallToAction:    input.RoleKnowledge.CallToAction,
			Goals:           sql.NullString{String: input.RoleKnowledge.Goals, Valid: input.RoleKnowledge.Goals != ""},
		})

		if err != nil {
			return err
		}

		member, err := tx.CreateBusinessMember(ctx, entity.CreateBusinessMemberParams{
			BusinessRootID: businessRootId,
			ProfileID:      input.ProfileID,
			Role:           entity.BusinessMemberRoleOwner,
			AnsweredAt:     sql.NullTime{Time: time.Now(), Valid: true},
			Status:         entity.BusinessMemberStatusAccepted,
		})

		if err != nil {
			return err
		}

		memberID = member.ID

		// TODO: send email notification to owner
		// TODO: log invitation to db

		return nil
	})

	if e != nil {
		return SetupBusinessRootFirstTimeResponse{}, errs.NewInternalServerError(e)
	}

	if memberID == 0 {
		return SetupBusinessRootFirstTimeResponse{}, errs.NewInternalServerError(errors.New("MEMBER_ID_IS_EMPTY"))
	}

	// REDIS: upsert one business into profile owned business cache
	err := s.obRepo.UpsertOneBusiness(ctx, input.ProfileID, owned_business_repository.RedisBusinessSub{
		BusinessRootID: businessRootId,
		Role:           entity.BusinessMemberRoleOwner,
		MemberID:       memberID,
	}, time.Hour)

	if err != nil {
		fmt.Println("redis upsert owned business failed:", err)
	}

	res := SetupBusinessRootFirstTimeResponse{
		ID: businessRootId,
	}

	return res, nil
}

func (s *BusinessInformationService) GetBusinessById(ctx context.Context, businessId int64, profileId uuid.UUID) (GetBusinessByIdResponse, error) {

	business, err := s.store.GetBusinessKnowledgeByBusinessRootID(ctx, businessId)

	if err == sql.ErrNoRows {
		return GetBusinessByIdResponse{}, errs.NewNotFound("BUSINESS_NOT_FOUND")
	}
	if err != nil && err != sql.ErrNoRows {
		return GetBusinessByIdResponse{}, err
	}

	members, err := s.store.GetMembersByBusinessRootID(ctx, businessId)
	if err != nil {
		return GetBusinessByIdResponse{}, err
	}

	var memberBusiness []BusinessMemberSub
	var userProfile *BusinessMemberSub
	for _, m := range members {
		if m.ProfileID == profileId {
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
		ID:             business.BusinessRootID,
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

func (s *BusinessInformationService) DeleteBusinessById(ctx context.Context, businessId int64, profileId uuid.UUID) (DeleteBusinessByIdResponse, error) {

	member, err := s.store.GetMemberByProfileIdAndBusinessRootId(ctx, entity.GetMemberByProfileIdAndBusinessRootIdParams{
		ProfileID:      profileId,
		BusinessRootID: businessId,
	})
	if err != nil {
		fmt.Println(err)
		return DeleteBusinessByIdResponse{}, err
	}

	if member.Role != entity.BusinessMemberRoleOwner {
		return DeleteBusinessByIdResponse{}, errs.NewForbidden("USER_NOT_OWNER")
	}

	root, err := s.store.GetBusinessRootById(ctx, businessId)

	if err == sql.ErrNoRows || root.DeletedAt.Valid {
		return DeleteBusinessByIdResponse{}, errs.NewNotFound("BUSINESS_NOT_FOUND")
	}

	if err != nil {
		return DeleteBusinessByIdResponse{}, err
	}

	members, err := s.store.GetMembersByBusinessRootID(ctx, businessId)
	if err != nil {
		return DeleteBusinessByIdResponse{}, err
	}

	e := s.store.ExecTx(ctx, func(tx *entity.Queries) error {
		rootId, err := tx.SoftDeleteBusinessRoot(ctx, businessId)
		if err != nil {
			return err
		}

		_, err = tx.SoftDeleteBusinessKnowledgeByBusinessRootID(ctx, rootId)
		if err != nil {
			return err
		}

		_, err = tx.SoftDeleteBusinessProductByBusinessRootID(ctx, rootId)
		if err != nil {
			return err
		}

		_, err = tx.SoftDeleteBusinessMemberByBusinessRootID(ctx, rootId)
		if err != nil {
			return err
		}

		_, err = tx.SoftDeleteBusinessRoleByBusinessRootID(ctx, rootId)
		if err != nil {
			return err
		}

		return nil
	})

	if e != nil {
		return DeleteBusinessByIdResponse{}, errs.NewInternalServerError(e)
	}

	// REDIS: delete cache to verify owned business to all members
	go func(members []entity.GetMembersByBusinessRootIDRow, businessID int64) {
		ctxBg, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		for _, m := range members {
			if err := s.obRepo.DeleteOneBusiness(ctxBg, m.ProfileID, businessID, time.Hour); err != nil {
				fmt.Println("redis delete cache failed:", err)
			}
		}
	}(members, businessId)

	return DeleteBusinessByIdResponse{
		ID: businessId,
	}, nil
}
