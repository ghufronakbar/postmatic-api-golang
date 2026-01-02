// internal/http/handler/affiliator_handler/referral_basic_handler.go
package affiliator_handler

import (
	"net/http"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/affiliator/referral_basic"

	"postmatic-api/pkg/response"

	"github.com/go-chi/chi/v5"
)

type ReferralBasicHandler struct {
	svc *referral_basic.ReferralBasicService
}

func NewReferralBasicHandler(referralSvc *referral_basic.ReferralBasicService) *ReferralBasicHandler {
	return &ReferralBasicHandler{svc: referralSvc}
}

func (h *ReferralBasicHandler) ReferralBasicRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetReferralBasicProfile)
	return r
}

func (h *ReferralBasicHandler) GetReferralBasicProfile(w http.ResponseWriter, r *http.Request) {
	prof, err := middleware.GetProfileFromContext(r.Context())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}
	res, err := h.svc.GetReferralBasicByProfileId(r.Context(), referral_basic.GetReferralBasicFilter{
		ProfileID: prof.ID,
	})
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_GET_REFERRAL_BASIC", res)
}
