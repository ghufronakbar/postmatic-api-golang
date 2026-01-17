// internal/module/app/referral_rule/handler/handler.go
package referral_rule_handler

import (
	"net/http"
	"postmatic-api/internal/internal_middleware"
	referral_rule_service "postmatic-api/internal/module/app/referral_rule/service"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	referralSvc *referral_rule_service.ReferralService
}

func NewHandler(referralSvc *referral_rule_service.ReferralService) *Handler {
	return &Handler{referralSvc: referralSvc}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetReferralRule)
	r.Post("/", h.UpsertReferralRule)

	return r
}

func (h *Handler) GetReferralRule(w http.ResponseWriter, r *http.Request) {
	res, err := h.referralSvc.GetRuleReferral(r.Context())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_GET_REFERRAL_RULE", res)
}

func (h *Handler) UpsertReferralRule(w http.ResponseWriter, r *http.Request) {

	var req referral_rule_service.UpsertAppProfileReferralRulesDTO
	prof, err := internal_middleware.GetProfileFromContext(r.Context())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	req.ProfileID = prof.ID

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.referralSvc.UpsertRuleReferral(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_UPSERT_REFERRAL_RULE", res)
}
