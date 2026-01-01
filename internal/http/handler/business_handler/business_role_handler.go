// internal/http/handler/business_handler/business_role_handler.go
package business_handler

import (
	"net/http"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/business/business_role"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type BusinessRoleHandler struct {
	busInSvc   *business_role.BusinessRoleService
	middleware *middleware.OwnedBusiness
}

func NewBusinessRoleHandler(busInSvc *business_role.BusinessRoleService, ownedMw *middleware.OwnedBusiness) *BusinessRoleHandler {
	return &BusinessRoleHandler{busInSvc: busInSvc, middleware: ownedMw}
}

func (h *BusinessRoleHandler) BusinessRoleRoutes() chi.Router {
	r := chi.NewRouter()

	// owned business middleware
	r.Route("/{businessId}", func(r chi.Router) {
		r.Use(h.middleware.OwnedBusinessMiddleware)
		r.Get("/", h.GetBusinessRoleByBusinessRootId)
		r.Post("/", h.UpsertBusinessRoleByBusinessRootID)
	})

	return r
}

func (h *BusinessRoleHandler) GetBusinessRoleByBusinessRootId(w http.ResponseWriter, r *http.Request) {

	business, err := middleware.OwnedBusinessFromContext(r.Context())

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

func (h *BusinessRoleHandler) UpsertBusinessRoleByBusinessRootID(w http.ResponseWriter, r *http.Request) {

	business, err := middleware.OwnedBusinessFromContext(r.Context())

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	var req business_role.UpsertBusinessRoleInput
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
