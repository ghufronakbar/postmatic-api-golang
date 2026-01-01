// internal/module/business/business_member/service.go
package business_member

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"postmatic-api/config"
	"postmatic-api/internal/module/headless/mailer"
	"postmatic-api/internal/module/headless/queue"
	"postmatic-api/internal/module/headless/token"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/internal/repository/redis/invitation_limiter_repository"
	"postmatic-api/internal/repository/redis/owned_business_repository"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/logger"
	"postmatic-api/pkg/pagination"
	"postmatic-api/pkg/utils"

	"github.com/google/uuid"
)

type BusinessMemberService struct {
	store   entity.Store
	cfg     config.Config
	queue   queue.MailerProducer
	token   *token.TokenMaker
	limiter *invitation_limiter_repository.LimiterInvitationRepo
	owned   *owned_business_repository.OwnedBusinessRepository
}

func NewService(store entity.Store, cfg config.Config, queue queue.MailerProducer, token *token.TokenMaker, limiter *invitation_limiter_repository.LimiterInvitationRepo, owned *owned_business_repository.OwnedBusinessRepository) *BusinessMemberService {
	return &BusinessMemberService{
		store:   store,
		cfg:     cfg,
		queue:   queue,
		token:   token,
		limiter: limiter,
		owned:   owned,
	}
}

// ================= CRUD =================

func (s *BusinessMemberService) GetBusinessMembersByBusinessRootID(ctx context.Context, filter GetBusinessMembersByBusinessRootIDFilter) ([]BusinessMemberResponse, pagination.Pagination, error) {
	members, err := s.store.GetMembersByBusinessRootIDWithStatus(ctx, entity.GetMembersByBusinessRootIDWithStatusParams{
		ProfileID:      filter.ProfileID,
		BusinessRootID: filter.BusinessRootID,
		IsVerified:     utils.NullBoolPtrToNullBool(filter.IsVerified),
	})
	if err != nil && err != sql.ErrNoRows {
		return nil, pagination.Pagination{}, err
	}

	now := time.Now()
	expDur := s.cfg.JWT_INVITATION_TOKEN_EXPIRED

	var expIds []int64
	var memberRes = make([]BusinessMemberResponse, 0, len(members))
	for _, member := range members {
		if member.Status == entity.BusinessMemberStatusPending &&
			!member.AnsweredAt.Valid &&
			now.After(member.UpdatedAt.Add(expDur)) {
			expIds = append(expIds, member.ID)
			member.Status = entity.BusinessMemberStatusExpired
		}
		var profileImage *string
		if member.ProfileImageUrl.Valid {
			profileImage = &member.ProfileImageUrl.String
		}
		var answeredAt *time.Time
		if member.AnsweredAt.Valid {
			answeredAt = &member.AnsweredAt.Time
		}
		memberRes = append(memberRes, BusinessMemberResponse{
			ID:     member.ID,
			Role:   string(member.Role),
			Status: string(member.Status),
			Profile: BusinessProfileSub{
				Name:  member.ProfileName,
				Email: member.ProfileEmail,
				Image: profileImage,
				ID:    member.ProfileID,
			},
			AnsweredAt: answeredAt,
			CreatedAt:  member.CreatedAt,
			UpdatedAt:  member.UpdatedAt,
			IsYourself: member.IsYourself,
		})
	}
	pag := pagination.NewPagination(&pagination.PaginationParams{
		Total: len(memberRes),
		Page:  1,
		Limit: len(memberRes),
	})

	// if have expired members, update status to expired in background
	if len(expIds) > 0 {
		go func(ids []int64) {
			ctxBg, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := s.store.UpdateManyBusinessMemberStatus(ctxBg, entity.UpdateManyBusinessMemberStatusParams{
				Status: entity.BusinessMemberStatusExpired,
				Ids:    ids,
			}); err != nil {
				logger.From(ctxBg).Error("Failed to update many business member status", "error", err)
			}
		}(append([]int64(nil), expIds...))
	}

	return memberRes, pag, nil
}

func (s *BusinessMemberService) InviteBusinessMember(ctx context.Context, input InviteBusinessMemberInput) (InviteMemberResponse, error) {
	checkMember, err := s.store.GetMemberByEmailAndBusinessRootId(ctx, entity.GetMemberByEmailAndBusinessRootIdParams{
		Email:          input.Email,
		BusinessRootID: input.BusinessRootID,
	})
	if err != nil && err != sql.ErrNoRows {
		return InviteMemberResponse{}, errs.NewInternalServerError(err)
	}

	var specialStatus []string = []string{
		string(entity.BusinessMemberStatusRejected),
		string(entity.BusinessMemberStatusLeft),
		string(entity.BusinessMemberStatusKicked),
	}

	isSpecialStatus := utils.StringInSlice(string(checkMember.Status), specialStatus)

	if checkMember.ID != 0 && !isSpecialStatus {
		return InviteMemberResponse{}, errs.NewBadRequest("MEMBER_ALREADY_EXISTS")
	}
	checkProf, err := s.store.GetProfileByEmail(ctx, input.Email)
	if err != nil && err != sql.ErrNoRows {
		return InviteMemberResponse{}, errs.NewInternalServerError(err)
	}

	checkBusinessRoot, err := s.store.GetBusinessKnowledgeByBusinessRootID(ctx, input.BusinessRootID)
	if err != nil && err != sql.ErrNoRows {
		return InviteMemberResponse{}, errs.NewInternalServerError(err)
	}

	if checkBusinessRoot.BusinessRootID == 0 {
		return InviteMemberResponse{}, errs.NewNotFound("BUSINESS_ROOT_NOT_FOUND")
	}

	var invMemRes InviteMemberResponse

	if isSpecialStatus && checkMember.ID != 0 {
		e := s.store.ExecTx(ctx, func(q *entity.Queries) error {
			mem, err := q.UpdateBusinessMemberStatus(ctx, entity.UpdateBusinessMemberStatusParams{
				ID:     checkMember.ID,
				Status: entity.BusinessMemberStatusPending,
			})
			if err != nil {
				return err
			}
			his, err := q.CreateBusinessMemberStatusHistory(ctx, entity.CreateBusinessMemberStatusHistoryParams{
				MemberID: mem.ID,
				Status:   entity.BusinessMemberStatusPending,
				Role:     entity.BusinessMemberRole(input.Role),
			})
			if err != nil {
				return err
			}
			link, err := s.createInvitationLink(createInvitationLinkInput{
				Email:                 input.Email,
				MemberID:              mem.ID,
				BusinessRootID:        checkMember.BusinessRootID,
				MemberRole:            string(mem.Role),
				MemberHistoryStatusID: his.ID,
			})
			if err != nil {
				return err
			}

			invMemRes = InviteMemberResponse{
				GeneralMemberResponse: GeneralMemberResponse{
					ID:             mem.ID,
					BusinessRootID: mem.BusinessRootID,
					Role:           string(mem.Role),
					Status:         string(mem.Status),
					CreatedAt:      mem.CreatedAt,
					UpdatedAt:      mem.UpdatedAt,
					ProfileID:      checkMember.ProfileID,
				},
				InvitationLink: link,
				RetryAfter:     s.cfg.CAN_RESEND_EMAIL_AFTER,
			}
			return nil
		})
		if e != nil {
			return InviteMemberResponse{}, e
		}
	} else if checkProf.ID != uuid.Nil && !isSpecialStatus {
		// direct invite with exisiting profile
		e := s.store.ExecTx(ctx, func(q *entity.Queries) error {
			mem, err := q.CreateBusinessMember(ctx, entity.CreateBusinessMemberParams{
				ProfileID:      checkProf.ID,
				BusinessRootID: input.BusinessRootID,
				Role:           entity.BusinessMemberRole(input.Role),
				Status:         entity.BusinessMemberStatusPending,
				AnsweredAt:     sql.NullTime{},
			})
			if err != nil {
				return err
			}
			his, err := q.CreateBusinessMemberStatusHistory(ctx, entity.CreateBusinessMemberStatusHistoryParams{
				MemberID: mem.ID,
				Status:   entity.BusinessMemberStatusPending,
				Role:     entity.BusinessMemberRole(input.Role),
			})
			if err != nil {
				return err
			}
			link, err := s.createInvitationLink(createInvitationLinkInput{
				MemberID:              mem.ID,
				BusinessRootID:        input.BusinessRootID,
				MemberRole:            string(mem.Role),
				MemberHistoryStatusID: his.ID,
				Email:                 input.Email,
			})
			if err != nil {
				return err
			}
			invMemRes = InviteMemberResponse{
				GeneralMemberResponse: GeneralMemberResponse{
					ID:             mem.ID,
					BusinessRootID: mem.BusinessRootID,
					Role:           string(mem.Role),
					Status:         string(mem.Status),
					CreatedAt:      mem.CreatedAt,
					UpdatedAt:      mem.UpdatedAt,
					ProfileID:      mem.ProfileID,
				},
				InvitationLink: link,
				RetryAfter:     s.cfg.CAN_RESEND_EMAIL_AFTER,
			}
			return nil
		})
		if e != nil {
			return InviteMemberResponse{}, errs.NewInternalServerError(e)
		}
	} else {
		// create profile then invite
		e := s.store.ExecTx(ctx, func(q *entity.Queries) error {
			profile, err := q.CreateProfile(ctx, entity.CreateProfileParams{
				Email:       input.Email,
				Name:        input.Email,
				ImageUrl:    sql.NullString{},
				CountryCode: sql.NullString{},
				Phone:       sql.NullString{},
				Description: sql.NullString{},
			})
			if err != nil {
				return err
			}
			mem, err := q.CreateBusinessMember(ctx, entity.CreateBusinessMemberParams{
				ProfileID:      profile.ID,
				BusinessRootID: input.BusinessRootID,
				Role:           entity.BusinessMemberRole(input.Role),
				Status:         entity.BusinessMemberStatusPending,
				AnsweredAt:     sql.NullTime{},
			})
			if err != nil {
				return err
			}
			his, err := q.CreateBusinessMemberStatusHistory(ctx, entity.CreateBusinessMemberStatusHistoryParams{
				MemberID: mem.ID,
				Status:   entity.BusinessMemberStatusPending,
				Role:     entity.BusinessMemberRole(input.Role),
			})
			if err != nil {
				return err
			}
			link, err := s.createInvitationLink(createInvitationLinkInput{
				MemberID:              mem.ID,
				BusinessRootID:        input.BusinessRootID,
				MemberRole:            string(mem.Role),
				MemberHistoryStatusID: his.ID,
				Email:                 input.Email,
			})
			if err != nil {
				return err
			}
			invMemRes = InviteMemberResponse{
				GeneralMemberResponse: GeneralMemberResponse{
					ID:             mem.ID,
					BusinessRootID: mem.BusinessRootID,
					Role:           string(mem.Role),
					Status:         string(mem.Status),
					CreatedAt:      mem.CreatedAt,
					UpdatedAt:      mem.UpdatedAt,
					ProfileID:      mem.ProfileID,
				},
				InvitationLink: link,
				RetryAfter:     s.cfg.CAN_RESEND_EMAIL_AFTER,
			}
			return nil
		})
		if e != nil {
			return InviteMemberResponse{}, errs.NewInternalServerError(e)
		}
	}

	_ = s.addEmailToQueue(mailer.MemberInvitationInputDTO{
		Email:        input.Email,
		ConfirmUrl:   invMemRes.InvitationLink,
		BusinessName: checkBusinessRoot.Name,
	}, checkBusinessRoot.BusinessRootID)

	return invMemRes, nil
}

func (s *BusinessMemberService) ResendMemberInvitation(ctx context.Context, input ResendEmailInvitationInput) (InviteMemberResponse, error) {
	checkMember, err := s.store.GetBusinessMemberStatusHistoryByMemberID(ctx, input.MemberID)
	if err != nil && err != sql.ErrNoRows {
		return InviteMemberResponse{}, errs.NewInternalServerError(err)
	}
	if checkMember.MemberID == 0 {
		return InviteMemberResponse{}, errs.NewBadRequest("MEMBER_NOT_FOUND")
	}

	if checkMember.BusinessRootID != input.BusinessRootID {
		return InviteMemberResponse{}, errs.NewBadRequest("MEMBER_NOT_FOUND")
	}

	disallowStatus := []string{
		string(entity.BusinessMemberStatusAccepted),
		string(entity.BusinessMemberStatusRejected),
		string(entity.BusinessMemberStatusLeft),
		string(entity.BusinessMemberStatusKicked),
	}

	if utils.StringInSlice(string(checkMember.MemberStatus), disallowStatus) {
		return InviteMemberResponse{}, errs.NewBadRequest("MEMBER_ALREADY_" + strings.ToUpper(string(checkMember.MemberStatus)))
	}

	remaining, err := s.checkEmailLimit(ctx, invitation_limiter_repository.LimiterInvitationInput{
		Email:          checkMember.ProfileEmail,
		BusinessRootID: input.BusinessRootID,
	})
	if err != nil {
		return InviteMemberResponse{
			RetryAfter: remaining,
			GeneralMemberResponse: GeneralMemberResponse{
				ID:             checkMember.MemberID,
				BusinessRootID: checkMember.MemberBusinessRootID,
				Role:           string(checkMember.MemberRole),
				Status:         string(checkMember.MemberStatus),
				CreatedAt:      checkMember.MemberCreatedAt,
				UpdatedAt:      checkMember.MemberUpdatedAt,
				ProfileID:      checkMember.MemberProfileID,
			},
		}, err
	}
	if remaining > 0 {
		return InviteMemberResponse{
			RetryAfter: remaining,
			GeneralMemberResponse: GeneralMemberResponse{
				ID:             checkMember.MemberID,
				BusinessRootID: checkMember.MemberBusinessRootID,
				Role:           string(checkMember.MemberRole),
				Status:         string(checkMember.MemberStatus),
				CreatedAt:      checkMember.MemberCreatedAt,
				UpdatedAt:      checkMember.MemberUpdatedAt,
				ProfileID:      checkMember.MemberProfileID,
			},
		}, errs.NewBadRequest("PLEASE_WAIT")
	}
	link, err := s.createInvitationLink(createInvitationLinkInput{
		MemberID:              checkMember.MemberID,
		BusinessRootID:        checkMember.MemberBusinessRootID,
		MemberRole:            string(checkMember.MemberRole),
		MemberHistoryStatusID: checkMember.HistoryMemberID,
		Email:                 checkMember.ProfileEmail,
	})
	if err != nil {
		return InviteMemberResponse{}, err
	}
	_ = s.addEmailToQueue(mailer.MemberInvitationInputDTO{
		Email:        checkMember.ProfileEmail,
		ConfirmUrl:   link,
		BusinessName: checkMember.BusinessRootName,
	}, checkMember.MemberBusinessRootID)
	return InviteMemberResponse{
		RetryAfter: s.cfg.CAN_RESEND_EMAIL_AFTER,
		GeneralMemberResponse: GeneralMemberResponse{
			ID:             checkMember.MemberID,
			BusinessRootID: checkMember.MemberBusinessRootID,
			Role:           string(checkMember.MemberRole),
			Status:         string(checkMember.MemberStatus),
			CreatedAt:      checkMember.MemberCreatedAt,
			UpdatedAt:      checkMember.MemberUpdatedAt,
			ProfileID:      checkMember.MemberProfileID,
		},
	}, nil
}

func (s *BusinessMemberService) EditMember(ctx context.Context, input EditMemberInput) (GeneralMemberResponse, error) {
	checkMember, err := s.store.GetBusinessMemberStatusHistoryByMemberID(ctx, input.MemberID)
	if err != nil {
		return GeneralMemberResponse{}, err
	}
	if checkMember.MemberID == 0 {
		return GeneralMemberResponse{}, errs.NewNotFound("MEMBER_NOT_FOUND")
	}
	if checkMember.MemberBusinessRootID != input.BusinessRootID {
		return GeneralMemberResponse{}, errs.NewNotFound("MEMBER_NOT_FOUND")
	}
	if checkMember.MemberRole == entity.BusinessMemberRoleOwner {
		return GeneralMemberResponse{}, errs.NewBadRequest("OWNER_CANNOT_EDITED")
	}

	if checkMember.MemberStatus != entity.BusinessMemberStatusAccepted {
		return GeneralMemberResponse{}, errs.NewBadRequest("MEMBER_ALREADY_" + strings.ToUpper(string(checkMember.MemberStatus)))
	}

	checkBusiness, err := s.store.GetBusinessKnowledgeByBusinessRootID(ctx, input.BusinessRootID)
	if err != nil && err != sql.ErrNoRows {
		return GeneralMemberResponse{}, err
	}
	if checkBusiness.BusinessRootID == 0 {
		return GeneralMemberResponse{}, errs.NewNotFound("BUSINESS_NOT_FOUND")
	}

	var memberRole entity.BusinessMemberRole
	if utils.StringInSlice(input.Role, []string{
		string(entity.BusinessMemberRoleOwner),
		string(entity.BusinessMemberRoleMember),
	}) {
		memberRole = entity.BusinessMemberRole(input.Role)
	} else {
		return GeneralMemberResponse{}, errs.NewBadRequest("INVALID_ROLE")
	}

	var res GeneralMemberResponse

	e := s.store.ExecTx(ctx, func(q *entity.Queries) error {
		upMem, err := q.UpdateBusinessMemberRole(ctx, entity.UpdateBusinessMemberRoleParams{
			Role: entity.BusinessMemberRole(memberRole),
			ID:   input.MemberID,
		})
		if err != nil {
			return err
		}
		res = GeneralMemberResponse{
			ID:             input.MemberID,
			BusinessRootID: checkMember.MemberBusinessRootID,
			Role:           string(upMem.Role),
			Status:         string(upMem.Status),
			CreatedAt:      upMem.CreatedAt,
			UpdatedAt:      upMem.UpdatedAt,
			ProfileID:      upMem.ProfileID,
		}
		_, err = q.CreateBusinessMemberStatusHistory(ctx, entity.CreateBusinessMemberStatusHistoryParams{
			MemberID: input.MemberID,
			Status:   entity.BusinessMemberStatusAccepted,
			Role:     entity.BusinessMemberRole(memberRole),
		})
		if err != nil {
			return err
		}
		return nil
	})
	if e != nil {
		return GeneralMemberResponse{}, e
	}

	// enqueue announce role email
	go func(email, businessName, newRole string) {
		ctxBg, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.queue.EnqueueAnnounceRole(ctxBg, mailer.MemberAnnounceRoleInputDTO{
			Email:        email,
			BusinessName: businessName,
			NewRole:      newRole,
		}); err != nil {
			logger.From(ctxBg).Error("failed to enqueue announce role email", "error", err)
		}
	}(checkMember.ProfileEmail, checkBusiness.Name, string(memberRole))

	// enqueue delete owned business
	go func(profileID uuid.UUID, businessRootID int64) {
		ctxBg, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.owned.DeleteOneBusiness(ctxBg, profileID, businessRootID, 5*time.Minute); err != nil {
			logger.From(ctxBg).Error("failed to delete owned business cache", "error", err)
		}
	}(checkMember.MemberProfileID, checkMember.BusinessRootID)

	return res, nil
}

func (s *BusinessMemberService) RemoveBusinessMember(ctx context.Context, input RemoveBusinessMemberInput) (GeneralMemberResponse, error) {
	checkMember, err := s.store.GetBusinessMemberStatusHistoryByMemberID(ctx, input.MemberID)
	if err != nil {
		return GeneralMemberResponse{}, err
	}
	if checkMember.MemberID == 0 {
		return GeneralMemberResponse{}, errs.NewNotFound("MEMBER_NOT_FOUND")
	}

	checkBusiness, err := s.store.GetBusinessKnowledgeByBusinessRootID(ctx, input.BusinessRootID)
	if err == sql.ErrNoRows {
		return GeneralMemberResponse{}, errs.NewNotFound("BUSINESS_NOT_FOUND")
	}
	if err != nil {
		return GeneralMemberResponse{}, err
	}
	if checkBusiness.BusinessRootID == 0 {
		return GeneralMemberResponse{}, errs.NewNotFound("BUSINESS_NOT_FOUND")
	}

	if checkMember.MemberBusinessRootID != input.BusinessRootID {
		return GeneralMemberResponse{}, errs.NewNotFound("MEMBER_NOT_FOUND")
	}

	if checkMember.MemberRole == entity.BusinessMemberRoleOwner {
		return GeneralMemberResponse{}, errs.NewBadRequest("OWNER_CANNOT_REMOVE_ITSELF")
	}

	disallowStatus := []string{
		string(entity.BusinessMemberStatusRejected),
		string(entity.BusinessMemberStatusLeft),
		string(entity.BusinessMemberStatusKicked),
		string(entity.BusinessMemberStatusExpired),
	}

	if utils.StringInSlice(string(checkMember.MemberStatus), disallowStatus) {
		return GeneralMemberResponse{}, errs.NewBadRequest("MEMBER_ALREADY_" + strings.ToUpper(string(checkMember.MemberStatus)))
	}

	var res GeneralMemberResponse

	e := s.store.ExecTx(ctx, func(q *entity.Queries) error {
		upMem, err := q.UpdateBusinessMemberStatus(ctx, entity.UpdateBusinessMemberStatusParams{
			Status: entity.BusinessMemberStatusKicked,
			ID:     input.MemberID,
		})
		if err != nil {
			return err
		}
		res = GeneralMemberResponse{
			ID:             input.MemberID,
			BusinessRootID: checkMember.MemberBusinessRootID,
			Role:           string(upMem.Role),
			Status:         string(upMem.Status),
			CreatedAt:      upMem.CreatedAt,
			UpdatedAt:      upMem.UpdatedAt,
			ProfileID:      upMem.ProfileID,
		}
		_, err = q.CreateBusinessMemberStatusHistory(ctx, entity.CreateBusinessMemberStatusHistoryParams{
			MemberID: input.MemberID,
			Status:   entity.BusinessMemberStatusKicked,
			Role:     checkMember.MemberRole,
		})
		if err != nil {
			return err
		}
		return nil
	})

	if e != nil {
		return GeneralMemberResponse{}, e
	}
	// enqueue announce kick email
	go func(email, businessName string) {
		ctxBg, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.queue.EnqueueAnnounceKick(ctxBg, mailer.MemberAnnounceKickInputDTO{
			Email:        email,
			BusinessName: businessName,
		}); err != nil {
			logger.From(ctxBg).Error("failed to enqueue announce kick email", "error", err)
		}
	}(checkMember.ProfileEmail, checkBusiness.Name)

	// enqueue delete owned business
	go func(profileID uuid.UUID, businessRootID int64) {
		ctxBg, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.owned.DeleteOneBusiness(ctxBg, profileID, businessRootID, 5*time.Minute); err != nil {
			logger.From(ctxBg).Error("failed to delete owned business cache", "error", err)
		}
	}(checkMember.MemberProfileID, checkMember.BusinessRootID)

	return res, nil
}

// ================= INVITATION =================

func (s *BusinessMemberService) VerifyMemberInvitation(ctx context.Context, input VerifyMemberInvitationInput) (BusinessMemberInvitationResponse, error) {
	// TODO: check link valid or not
	// TODO: decoded
	return BusinessMemberInvitationResponse{}, nil
}

func (s *BusinessMemberService) AnswerMemberInvitation(ctx context.Context, input AnswerMemberInvitationInput) (GeneralMemberResponse, error) {
	// TODO: check link valid or not with VerifyMemberInvitation
	// TODO: decoded
	// TODO: update database
	return GeneralMemberResponse{}, nil
}

// HELPER

type createInvitationLinkInput struct {
	MemberID              int64
	BusinessRootID        int64
	MemberRole            string
	MemberHistoryStatusID int64
	Email                 string
}

func (s *BusinessMemberService) createInvitationLink(input createInvitationLinkInput) (string, error) {
	token, err := s.token.GenerateInvitationToken(token.GenerateInvitationTokenInput{
		ID:                    uuid.New(),
		MemberID:              input.MemberID,
		BusinessRootID:        input.BusinessRootID,
		MemberRole:            input.MemberRole,
		MemberHistoryStatusID: input.MemberHistoryStatusID,
	})
	if err != nil {
		return "", err
	}
	base := strings.TrimRight(s.cfg.AUTH_URL, "/")
	route := strings.TrimLeft(s.cfg.INVITE_MEMBER_ROUTE, "/")
	link := fmt.Sprintf("%s/%s/%s", base, route, token)
	return link, nil
}

func (s *BusinessMemberService) checkEmailLimit(ctx context.Context, input invitation_limiter_repository.LimiterInvitationInput) (int64, error) {
	check, err := s.limiter.GetLimiterInvitation(ctx, input)
	if err != nil {
		return 0, err
	}
	if check == nil {
		return 0, nil
	}
	return check.RetryAfterSeconds, nil
}

func (s *BusinessMemberService) addEmailToQueue(input mailer.MemberInvitationInputDTO, businessRootID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.limiter.SaveLimiterInvitation(ctx, invitation_limiter_repository.LimiterInvitationInput{
		Email:          input.Email,
		BusinessRootID: businessRootID,
	}, time.Second*time.Duration(s.cfg.CAN_RESEND_EMAIL_AFTER))
	if err != nil {
		logger.From(ctx).Error("failed to save limiter invitation", "error", err)
	}
	err = s.queue.EnqueueInvitation(ctx, input)
	if err != nil {
		return err
	}
	return nil
}
