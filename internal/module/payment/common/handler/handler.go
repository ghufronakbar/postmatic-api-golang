// internal/module/payment/common/handler/handler.go
package payment_common_handler

import (
	"net/http"

	"postmatic-api/internal/internal_middleware"
	payment_common_service "postmatic-api/internal/module/payment/common/service"
	"postmatic-api/pkg/filter"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type PaymentCommonHandler struct {
	service    *payment_common_service.PaymentCommonService
	middleware *internal_middleware.OwnedBusiness
}

func NewHandler(service *payment_common_service.PaymentCommonService, middleware *internal_middleware.OwnedBusiness) *PaymentCommonHandler {
	return &PaymentCommonHandler{
		service:    service,
		middleware: middleware,
	}
}

func (h *PaymentCommonHandler) Routes(allAllowedMiddleware func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(allAllowedMiddleware)

	// Profile-based routes
	r.Get("/profile", h.GetPaymentHistoriesByProfile)

	// Business-based routes with OwnedBusiness middleware
	r.Route("/{businessId}", func(r chi.Router) {
		r.Use(h.middleware.OwnedBusinessMiddleware)
		r.Use(func(next http.Handler) http.Handler {
			return internal_middleware.ReqFilterMiddleware(next, payment_common_service.SORT_BY)
		})
		r.Get("/", h.GetPaymentHistoriesByBusiness)
		r.Get("/{id}", h.GetPaymentHistoryByIdAndBusiness)
		r.Post("/{id}/cancel", h.CancelPaymentByBusiness)
	})

	return r
}

// WebhookRoute returns the handler for webhook (no auth)
func (h *PaymentCommonHandler) WebhookRoute() http.HandlerFunc {
	return h.HandleWebhook
}

// GetPaymentHistoriesByProfile godoc
// @Summary Get payment histories by profile
// @Tags Payment
// @Accept json
// @Produce json
// @Param search query string false "Search by product name or payment method"
// @Param status query string false "Filter by status (pending, success, failed, etc)"
// @Param sortBy query string false "Sort by field (created_at, total_amount)"
// @Param sortDir query string false "Sort direction (asc, desc)"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Response{data=[]payment_common_service.PaymentHistoryResponse}
// @Router /api/payment/profile [get]
func (h *PaymentCommonHandler) GetPaymentHistoriesByProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get profile from context (ID is the Profile ID)
	claims, _ := internal_middleware.GetProfileFromContext(ctx)
	if claims == nil {
		response.ValidationFailed(w, r, map[string]string{"authorization": "PROFILE_ID_REQUIRED"})
		return
	}

	// Parse query params manually for profile route (no middleware)
	status := r.URL.Query().Get("status")
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	search := r.URL.Query().Get("search")
	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	svcFilter := payment_common_service.GetPaymentHistoriesFilter{
		ProfileID: claims.ID.String(),
		Search:    searchPtr,
		Status:    statusPtr,
		SortBy:    getQueryOr(r, "sortBy", "created_at"),
		SortDir:   getQueryOr(r, "sort", "desc"),
		Page:      getQueryIntOr(r, "page", 1),
		Limit:     getQueryIntOr(r, "limit", 10),
	}

	data, pag, err := h.service.GetPaymentHistoriesByProfile(ctx, svcFilter)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	reqFilter := &filter.ReqFilter{
		Search: search,
		Page:   svcFilter.Page,
		Limit:  svcFilter.Limit,
		SortBy: svcFilter.SortBy,
		Sort:   svcFilter.SortDir,
	}

	response.LIST(w, r, "PAYMENT_HISTORIES_RETRIEVED", data, reqFilter, pag)
}

// GetPaymentHistoriesByBusiness godoc
// @Summary Get payment histories by business
// @Tags Payment
// @Accept json
// @Produce json
// @Param businessId path int true "Business Root ID"
// @Param search query string false "Search by product name or payment method"
// @Param status query string false "Filter by status (pending, success, failed, etc)"
// @Param sortBy query string false "Sort by field (created_at, total_amount)"
// @Param sort query string false "Sort direction (asc, desc)"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Response{data=[]payment_common_service.PaymentHistoryResponse}
// @Router /api/payment/{businessId} [get]
func (h *PaymentCommonHandler) GetPaymentHistoriesByBusiness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get business context from middleware
	ownedBusiness, err := internal_middleware.OwnedBusinessFromContext(ctx)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	// Get filter from middleware
	reqFilter := internal_middleware.GetFilterFromContext(ctx)

	// Parse status
	status := r.URL.Query().Get("status")
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	// Parse search
	var searchPtr *string
	if reqFilter.Search != "" {
		searchPtr = &reqFilter.Search
	}

	svcFilter := payment_common_service.GetPaymentHistoriesByBusinessFilter{
		BusinessRootID: ownedBusiness.BusinessRootID,
		Search:         searchPtr,
		Status:         statusPtr,
		SortBy:         reqFilter.SortBy,
		SortDir:        reqFilter.Sort,
		Page:           reqFilter.Page,
		Limit:          reqFilter.Limit,
	}

	data, pag, err := h.service.GetPaymentHistoriesByBusiness(ctx, svcFilter)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.LIST(w, r, "PAYMENT_HISTORIES_RETRIEVED", data, &reqFilter, pag)
}

// GetPaymentHistoryByIdAndBusiness godoc
// @Summary Get payment history by ID and business
// @Tags Payment
// @Accept json
// @Produce json
// @Param businessId path int true "Business Root ID"
// @Param id path string true "Payment ID"
// @Success 200 {object} response.Response{data=payment_common_service.PaymentHistoryResponse}
// @Router /api/payment/{businessId}/{id} [get]
func (h *PaymentCommonHandler) GetPaymentHistoryByIdAndBusiness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	if id == "" {
		response.ValidationFailed(w, r, map[string]string{"id": "REQUIRED"})
		return
	}

	// Get business context from middleware
	ownedBusiness, err := internal_middleware.OwnedBusinessFromContext(ctx)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	data, err := h.service.GetPaymentHistoryByIdAndBusiness(ctx, id, ownedBusiness.BusinessRootID)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "PAYMENT_HISTORY_RETRIEVED", data)
}

// CancelPaymentByBusiness godoc
// @Summary Cancel a pending payment by business
// @Tags Payment
// @Accept json
// @Produce json
// @Param businessId path int true "Business Root ID"
// @Param id path string true "Payment ID"
// @Success 200 {object} response.Response{data=payment_common_service.PaymentHistoryResponse}
// @Router /api/payment/{businessId}/{id}/cancel [post]
func (h *PaymentCommonHandler) CancelPaymentByBusiness(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	if id == "" {
		response.ValidationFailed(w, r, map[string]string{"id": "REQUIRED"})
		return
	}

	// Get business context from middleware
	ownedBusiness, err := internal_middleware.OwnedBusinessFromContext(ctx)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	data, err := h.service.CancelPaymentByBusiness(ctx, id, ownedBusiness.BusinessRootID)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "PAYMENT_CANCELED", data)
}

// HandleWebhook godoc
// @Summary Handle Midtrans webhook
// @Tags Payment
// @Accept json
// @Produce json
// @Param body body payment_common_service.MidtransNotification true "Midtrans notification"
// @Success 200 {object} response.Response
// @Router /api/payment/webhook [post]
func (h *PaymentCommonHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var notification payment_common_service.MidtransNotification

	// Validate request body with utils.ValidateStruct
	if appErr := utils.ValidateStruct(r.Body, &notification); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	err := h.service.HandleWebhook(ctx, notification)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "WEBHOOK_PROCESSED", nil)
}

// Helper functions
func getQueryOr(r *http.Request, key, defaultVal string) string {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func getQueryIntOr(r *http.Request, key string, defaultVal int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	intVal := 0
	for _, c := range val {
		if c >= '0' && c <= '9' {
			intVal = intVal*10 + int(c-'0')
		} else {
			return defaultVal
		}
	}
	if intVal <= 0 {
		return defaultVal
	}
	return intVal
}
