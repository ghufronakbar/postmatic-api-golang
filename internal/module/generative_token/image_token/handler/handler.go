// internal/module/generative_token/image_token/handler/handler.go
package image_token_handler

import (
	"net/http"

	"postmatic-api/internal/internal_middleware"
	image_token_service "postmatic-api/internal/module/generative_token/image_token/service"
	"postmatic-api/pkg/filter"
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
		r.Use(func(next http.Handler) http.Handler {
			return internal_middleware.ReqFilterMiddleware(next, image_token_service.SORT_BY)
		})
		r.Get("/", h.GetTokenTransactions)
		r.Get("/status", h.GetTokenStatus)
	})

	return r
}

// GetTokenTransactions handles GET /api/app/generative-token/{businessId}/image-token/
// @Summary Get token transaction history
// @Tags GenerativeToken
// @Accept json
// @Produce json
// @Param businessId path int true "Business Root ID"
// @Param category query string false "Filter by type (in, out)"
// @Param sortBy query string false "Sort by field (created_at, amount)"
// @Param sort query string false "Sort direction (asc, desc)"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Response{data=[]image_token_service.TokenTransactionResponse}
// @Router /api/app/generative-token/{businessId}/image-token [get]
func (h *Handler) GetTokenTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get business context from middleware
	ownedBusiness, err := internal_middleware.OwnedBusinessFromContext(ctx)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	// Get filter from middleware
	reqFilter := internal_middleware.GetFilterFromContext(ctx)

	// Parse category (in/out) from query
	var typeFilter *string
	category := reqFilter.Category
	if category == "in" || category == "out" {
		typeFilter = &category
	}

	svcFilter := image_token_service.GetTokenTransactionsFilter{
		BusinessRootID: ownedBusiness.BusinessRootID,
		Type:           typeFilter,
		DateStart:      reqFilter.DateStart,
		DateEnd:        reqFilter.DateEnd,
		SortBy:         reqFilter.SortBy,
		SortDir:        reqFilter.Sort,
		Page:           reqFilter.Page,
		Limit:          reqFilter.Limit,
	}

	data, pag, err := h.svc.GetTokenTransactions(ctx, svcFilter)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "TOKEN_TRANSACTIONS_RETRIEVED", data, &filter.ReqFilter{
		Search:   reqFilter.Search,
		Page:     reqFilter.Page,
		Limit:    reqFilter.Limit,
		SortBy:   reqFilter.SortBy,
		Sort:     reqFilter.Sort,
		Category: reqFilter.Category,
	}, pag)
}

// GetTokenStatus handles GET /api/app/generative-token/{businessId}/image-token/status
// @Summary Get token status
// @Tags GenerativeToken
// @Accept json
// @Produce json
// @Param businessId path int true "Business Root ID"
// @Success 200 {object} response.Response{data=image_token_service.TokenStatusResponse}
// @Router /api/app/generative-token/{businessId}/image-token/status [get]
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

	response.OK(w, r, "TOKEN_STATUS_RETRIEVED", result)
}
