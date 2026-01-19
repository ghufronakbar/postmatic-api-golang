// internal/module/generative_token/image_token/handler/handler.go
package image_token_handler

import (
	"net/http"

	"postmatic-api/internal/internal_middleware"
	image_token_service "postmatic-api/internal/module/generative_token/image_token/service"
	"postmatic-api/pkg/response"

	"github.com/go-chi/chi/v5"
)

// Handler handles generative token image token HTTP requests
type Handler struct {
	svc        *image_token_service.ImageTokenService
	middleware *internal_middleware.OwnedBusiness
}

// NewHandler creates a new Handler
func NewHandler(svc *image_token_service.ImageTokenService, middleware *internal_middleware.OwnedBusiness) *Handler {
	return &Handler{
		svc:        svc,
		middleware: middleware,
	}
}

// Routes returns the routes for generative token image token
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Route("/{businessId}", func(r chi.Router) {
		r.Use(h.middleware.OwnedBusinessMiddleware)
		r.Get("/status", h.GetTokenStatus)
	})

	return r
}

// GetTokenStatus handles GET /api/app/generative-token/{businessId}/image-token/status
func (h *Handler) GetTokenStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get business context from middleware
	ownedBusiness, err := internal_middleware.OwnedBusinessFromContext(ctx)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	result, err := h.svc.GetTokenStatus(ctx, ownedBusiness.BusinessRootID)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS", result)
}
