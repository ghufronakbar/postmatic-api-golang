// internal/http/handler/business_handler/business_information_handler.go
package business_handler

import (
	"net/http"
	"postmatic-api/internal/http/middleware"
	"postmatic-api/internal/module/business/business_knowledge"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type BusinessKnowledgeHandler struct {
	busInSvc   *business_knowledge.BusinessKnowledgeService
	middleware *middleware.OwnedBusiness
}

func NewBusinessKnowledgeHandler(busInSvc *business_knowledge.BusinessKnowledgeService, ownedMw *middleware.OwnedBusiness) *BusinessKnowledgeHandler {
	return &BusinessKnowledgeHandler{busInSvc: busInSvc, middleware: ownedMw}
}

func (h *BusinessKnowledgeHandler) BusinessKnowledgeRoutes() chi.Router {
	r := chi.NewRouter()

	// owned business middleware
	r.Route("/{businessId}", func(r chi.Router) {
		r.Use(h.middleware.OwnedBusinessMiddleware)
		r.Get("/", h.GetBusinessKnowledgeByBusinessRootId)
		r.Post("/", h.UpsertBusinessKnowledgeByBusinessRootID)
	})

	return r
}

func (h *BusinessKnowledgeHandler) GetBusinessKnowledgeByBusinessRootId(w http.ResponseWriter, r *http.Request) {

	business, err := middleware.OwnedBusinessFromContext(r.Context())

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

func (h *BusinessKnowledgeHandler) UpsertBusinessKnowledgeByBusinessRootID(w http.ResponseWriter, r *http.Request) {
	businessId := chi.URLParam(r, "businessId")

	var req business_knowledge.UpsertBusinessKnowledgeInput

	if appErr := utils.ValidateStruct(r.Body, &req); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	res, err := h.busInSvc.UpsertBusinessKnowledgeByBusinessRootID(r.Context(), businessId, req)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_UPDATE_BUSINESS_KNOWLEDGE", res)
}
