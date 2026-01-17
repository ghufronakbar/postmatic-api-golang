// internal/module/business/business_timezone_pref/handler/handler.go
package business_timezone_pref_handler

import (
	"net/http"
	"postmatic-api/internal/internal_middleware"
	business_timezone_pref_service "postmatic-api/internal/module/business/business_timezone_pref/service"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	tzSvc      *business_timezone_pref_service.BusinessTimezonePrefService
	middleware *internal_middleware.OwnedBusiness
}

func NewHandler(tzSvc *business_timezone_pref_service.BusinessTimezonePrefService, ownedMw *internal_middleware.OwnedBusiness) *Handler {
	return &Handler{tzSvc: tzSvc, middleware: ownedMw}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// owned business middleware
	r.Route("/{businessId}", func(r chi.Router) {
		r.Use(h.middleware.OwnedBusinessMiddleware)
		r.Get("/", h.GetBusinessTimezonePrefByBusinessRootId)
		r.Post("/", h.UpsertBusinessTimezonePrefByBusinessRootID)
	})

	return r
}

func (h *Handler) GetBusinessTimezonePrefByBusinessRootId(w http.ResponseWriter, r *http.Request) {

	business, err := internal_middleware.OwnedBusinessFromContext(r.Context())

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

func (h *Handler) UpsertBusinessTimezonePrefByBusinessRootID(w http.ResponseWriter, r *http.Request) {
	business, err := internal_middleware.OwnedBusinessFromContext(r.Context())

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	var req business_timezone_pref_service.UpsertBusinessTimezonePrefInput
	req.BusinessRootID = business.BusinessRootID

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.tzSvc.UpsertBusinessTimezonePrefByBusinessRootID(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_UPDATE_BUSINESS_TIMEZONE_PREF", res)
}
