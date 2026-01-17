// internal/module/affiliator/referral_basic/handler/handler.go
package referral_basic_handler

import (
	"net/http"
	"postmatic-api/internal/internal_middleware"
	referral_basic_service "postmatic-api/internal/module/affiliator/referral_basic/service"

	"postmatic-api/pkg/response"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *referral_basic_service.ReferralBasicService
}

func NewHandler(referralSvc *referral_basic_service.ReferralBasicService) *Handler {
	return &Handler{svc: referralSvc}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetReferralBasicProfile)
	return r
}

func (h *Handler) GetReferralBasicProfile(w http.ResponseWriter, r *http.Request) {
	prof, err := internal_middleware.GetProfileFromContext(r.Context())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}
	res, err := h.svc.GetReferralBasicByProfileId(r.Context(), referral_basic_service.GetReferralBasicFilter{
		ProfileID: prof.ID,
	})
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_GET_REFERRAL_BASIC", res)
}
