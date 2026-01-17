// internal/module/business/business_member/handler/handler.go
package business_member_handler

import (
	"net/http"
	"postmatic-api/internal/internal_middleware"
	business_member_service "postmatic-api/internal/module/business/business_member/service"
	"postmatic-api/internal/repository/entity"
	"strconv"

	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	busInSvc   *business_member_service.BusinessMemberService
	middleware *internal_middleware.OwnedBusiness
}

func NewHandler(busInSvc *business_member_service.BusinessMemberService, ownedMw *internal_middleware.OwnedBusiness) *Handler {
	return &Handler{busInSvc: busInSvc, middleware: ownedMw}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// owned business middleware
	r.Route("/{businessId}", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(h.middleware.OwnedBusinessMiddleware)

			r.Get("/", h.GetMembersByBusinessID)
			r.Post("/", h.InviteBusinessMember)
			r.Post("/resend-invitation", h.ResendMemberInvitation)
			r.Put("/", h.EditMember)
		})

		r.Route("/{memberInvitationTokenOrMemberId}", func(r chi.Router) {
			r.Route("/", func(r chi.Router) {
				r.Use(h.middleware.OwnedBusinessMiddleware)
				r.Delete("/", h.RemoveMember)
			})
			r.Get("/verify", h.VerifyMemberInvitation)
			r.Post("/answer", h.AnswerMemberInvitation)
		})

	})

	return r
}

func (h *Handler) GetMembersByBusinessID(w http.ResponseWriter, r *http.Request) {
	prof, _ := internal_middleware.GetProfileFromContext(r.Context())
	business, _ := internal_middleware.OwnedBusinessFromContext(r.Context())

	filter := internal_middleware.GetFilterFromContext(r.Context())

	var verified *bool
	if filter.Category != "" {
		switch filter.Category {
		case "verified":
			verified = utils.ParseBoolPtr("true")
		case "unverified":
			verified = utils.ParseBoolPtr("false")
		default:
			verified = nil
		}
	}

	fquery := business_member_service.GetBusinessMembersByBusinessRootIDFilter{
		BusinessRootID: business.BusinessRootID,
		ProfileID:      prof.ID,
		IsVerified:     verified,
	}

	res, pag, err := h.busInSvc.GetBusinessMembersByBusinessRootID(r.Context(), fquery)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "SUCCESS_GET_BUSINESS_MEMBERS", res, &filter, &pag)
}

func (h *Handler) InviteBusinessMember(w http.ResponseWriter, r *http.Request) {
	var req business_member_service.InviteBusinessMemberInput

	business, _ := internal_middleware.OwnedBusinessFromContext(r.Context())
	req.BusinessRootID = business.BusinessRootID

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.busInSvc.InviteBusinessMember(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_INVITE_BUSINESS_MEMBER", res)
}

func (h *Handler) EditMember(w http.ResponseWriter, r *http.Request) {
	var req business_member_service.EditMemberInput

	prof, _ := internal_middleware.GetProfileFromContext(r.Context())
	business, _ := internal_middleware.OwnedBusinessFromContext(r.Context())
	req.BusinessRootID = business.BusinessRootID
	req.ProfileID = prof.ID

	if business.Role != entity.BusinessMemberRoleOwner {
		response.Error(w, r, errs.NewForbidden("ONLY_OWNER_CAN_ACCESS"), nil)
		return
	}

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.busInSvc.EditMember(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_EDIT_BUSINESS_MEMBER", res)
}

func (h *Handler) ResendMemberInvitation(w http.ResponseWriter, r *http.Request) {
	var req business_member_service.ResendEmailInvitationInput

	prof, _ := internal_middleware.GetProfileFromContext(r.Context())
	buss, _ := internal_middleware.OwnedBusinessFromContext(r.Context())
	req.ProfileID = prof.ID
	req.BusinessRootID = buss.BusinessRootID

	if buss.Role != entity.BusinessMemberRoleOwner {
		response.Error(w, r, errs.NewForbidden("ONLY_OWNER_CAN_ACCESS"), nil)
		return
	}

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.busInSvc.ResendMemberInvitation(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, res)
		return
	}

	response.OK(w, r, "SUCCESS_RESEND_BUSINESS_MEMBER_INVITATION", res)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	var req business_member_service.RemoveBusinessMemberInput

	prof, _ := internal_middleware.GetProfileFromContext(r.Context())
	buss, _ := internal_middleware.OwnedBusinessFromContext(r.Context())
	req.ProfileID = prof.ID
	req.BusinessRootID = buss.BusinessRootID

	intMemberID, err := strconv.ParseInt(chi.URLParam(r, "memberInvitationTokenOrMemberId"), 10, 64)
	if err != nil {
		response.Error(w, r, errs.NewValidationFailed(map[string]string{
			"memberId": "memberId must be an integer64",
		}), nil)
		return
	}
	req.MemberID = intMemberID

	if buss.Role != entity.BusinessMemberRoleOwner {
		response.Error(w, r, errs.NewForbidden("ONLY_OWNER_CAN_ACCESS"), nil)
		return
	}

	res, err := h.busInSvc.RemoveBusinessMember(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_REMOVE_BUSINESS_MEMBER", res)
}

// INVITATION

func (h *Handler) VerifyMemberInvitation(w http.ResponseWriter, r *http.Request) {
	var req business_member_service.VerifyMemberInvitationInput

	req.MemberInvitationToken = chi.URLParam(r, "memberInvitationTokenOrMemberId")

	res, err := h.busInSvc.VerifyMemberInvitation(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, res)
		return
	}

	response.OK(w, r, "SUCCESS_VERIFY_BUSINESS_MEMBER", res)
}

func (h *Handler) AnswerMemberInvitation(w http.ResponseWriter, r *http.Request) {
	var req business_member_service.AnswerMemberInvitationInput

	req.MemberInvitationToken = chi.URLParam(r, "memberInvitationTokenOrMemberId")

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.busInSvc.AnswerMemberInvitation(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, res)
		return
	}

	response.OK(w, r, "SUCCESS_ANSWER_BUSINESS_MEMBER_INVITATION", res)
}
