// internal/module/business/business_knowledge/handler/handler.go
package business_knowledge_handler

import (
	"net/http"
	"postmatic-api/internal/internal_middleware"
	business_knowledge_service "postmatic-api/internal/module/business/business_knowledge/service"
	"strconv"

	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	busInSvc   *business_knowledge_service.BusinessKnowledgeService
	middleware *internal_middleware.OwnedBusiness
}

func NewHandler(busInSvc *business_knowledge_service.BusinessKnowledgeService, ownedMw *internal_middleware.OwnedBusiness) *Handler {
	return &Handler{busInSvc: busInSvc, middleware: ownedMw}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	// owned business middleware
	r.Route("/{businessId}", func(r chi.Router) {
		r.Use(h.middleware.OwnedBusinessMiddleware)
		r.Get("/", h.GetBusinessKnowledgeByBusinessRootId)
		r.Post("/", h.UpsertBusinessKnowledgeByBusinessRootID)
	})

	return r
}

func (h *Handler) GetBusinessKnowledgeByBusinessRootId(w http.ResponseWriter, r *http.Request) {

	business, err := internal_middleware.OwnedBusinessFromContext(r.Context())

	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	res, err := h.busInSvc.GetBusinessKnowledgeByBusinessRootID(r.Context(), business.BusinessRootID)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_GET_BUSINESS_KNOWLEDGE", res)
}

func (h *Handler) UpsertBusinessKnowledgeByBusinessRootID(w http.ResponseWriter, r *http.Request) {
	businessId := chi.URLParam(r, "businessId")

	intBusinessId, err := strconv.ParseInt(businessId, 10, 64)
	if err != nil {
		response.Error(w, r, errs.NewValidationFailed(map[string]string{
			"businessId": "businessId must be an integer64",
		}), nil)
		return
	}

	var req business_knowledge_service.UpsertBusinessKnowledgeInput

	req.BusinessRootID = intBusinessId

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.busInSvc.UpsertBusinessKnowledgeByBusinessRootID(r.Context(), req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_UPDATE_BUSINESS_KNOWLEDGE", res)
}
