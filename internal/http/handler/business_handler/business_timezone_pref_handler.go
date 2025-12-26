// internal/http/handler/business_handler/business_role_handler.go
package business_handler

import (
	"net/http"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/business/business_timezone_pref"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type BusinessTimezonePrefHandler struct {
	tzSvc      *business_timezone_pref.BusinessTimezonePrefService
	middleware *middleware.OwnedBusiness
}

func NewBusinessTimezonePrefHandler(tzSvc *business_timezone_pref.BusinessTimezonePrefService, ownedMw *middleware.OwnedBusiness) *BusinessTimezonePrefHandler {
	return &BusinessTimezonePrefHandler{tzSvc: tzSvc, middleware: ownedMw}
}

func (h *BusinessTimezonePrefHandler) BusinessTimezonePrefRoutes() chi.Router {
	r := chi.NewRouter()

	// owned business middleware
	r.Route("/{businessId}", func(r chi.Router) {
		r.Use(h.middleware.OwnedBusinessMiddleware)
		r.Get("/", h.GetBusinessTimezonePrefByBusinessRootId)
		r.Post("/", h.UpsertBusinessTimezonePrefByBusinessRootID)
	})

	return r
}

func (h *BusinessTimezonePrefHandler) GetBusinessTimezonePrefByBusinessRootId(w http.ResponseWriter, r *http.Request) {

	business, err := middleware.OwnedBusinessFromContext(r.Context())

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	res, err := h.tzSvc.GetBusinessTimezonePrefByBusinessRootID(r.Context(), business.BusinessRootID)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_GET_BUSINESS_TIMEZONE_PREF", res)
}

func (h *BusinessTimezonePrefHandler) UpsertBusinessTimezonePrefByBusinessRootID(w http.ResponseWriter, r *http.Request) {
	businessId := chi.URLParam(r, "businessId")

	var req business_timezone_pref.UpsertBusinessTimezonePrefInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.tzSvc.UpsertBusinessTimezonePrefByBusinessRootID(r.Context(), businessId, req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_UPDATE_BUSINESS_TIMEZONE_PREF", res)
}
