// internal/http/handler/app_handler/referral_rule_handler.go
package app_handler

import (
	"net/http"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/app/referral"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type ReferralRuleHandler struct {
	referralSvc *referral.ReferralService
}

func NewReferralRuleHandler(referralSvc *referral.ReferralService) *ReferralRuleHandler {
	return &ReferralRuleHandler{referralSvc: referralSvc}
}

func (h *ReferralRuleHandler) ReferralRuleRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetReferralRule)
	r.Post("/", h.UpsertReferralRule)

	return r
}

func (h *ReferralRuleHandler) GetReferralRule(w http.ResponseWriter, r *http.Request) {
	res, err := h.referralSvc.GetRuleReferral(r.Context())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_GET_REFERRAL_RULE", res)
}

func (h *ReferralRuleHandler) UpsertReferralRule(w http.ResponseWriter, r *http.Request) {

	var req referral.UpsertAppProfileReferralRulesDTO
	prof, err := middleware.GetUserFromContext(r.Context())
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
