// internal/module/business/business_role/handler/handler.go
package business_role_handler

import (
	"net/http"
	"postmatic-api/internal/internal_middleware"
	business_role_service "postmatic-api/internal/module/business/business_role/service"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	busInSvc   *business_role_service.BusinessRoleService
	middleware *internal_middleware.OwnedBusiness
}

func NewHandler(busInSvc *business_role_service.BusinessRoleService, ownedMw *internal_middleware.OwnedBusiness) *Handler {
	return &Handler{busInSvc: busInSvc, middleware: ownedMw}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// owned business middleware
	r.Route("/{businessId}", func(r chi.Router) {
		r.Use(h.middleware.OwnedBusinessMiddleware)
		r.Get("/", h.GetBusinessRoleByBusinessRootId)
		r.Post("/", h.UpsertBusinessRoleByBusinessRootID)
	})

	return r
}

func (h *Handler) GetBusinessRoleByBusinessRootId(w http.ResponseWriter, r *http.Request) {

	business, err := internal_middleware.OwnedBusinessFromContext(r.Context())

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	res, err := h.busInSvc.GetBusinessRoleByBusinessRootID(r.Context(), business.BusinessRootID)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_GET_BUSINESS_ROLE", res)
}

func (h *Handler) UpsertBusinessRoleByBusinessRootID(w http.ResponseWriter, r *http.Request) {

	business, err := internal_middleware.OwnedBusinessFromContext(r.Context())

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	var req business_role_service.UpsertBusinessRoleInput
	req.BusinessRootID = business.BusinessRootID

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.busInSvc.UpsertBusinessRoleByBusinessRootID(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_UPDATE_BUSINESS_ROLE", res)
}
