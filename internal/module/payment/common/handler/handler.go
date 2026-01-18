// internal/module/payment/common/handler/handler.go
package payment_common_handler

import (
	"net/http"
	"strconv"

	"postmatic-api/internal/internal_middleware"
	payment_common_service "postmatic-api/internal/module/payment/common/service"
	"postmatic-api/pkg/filter"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type PaymentCommonHandler struct {
	service *payment_common_service.PaymentCommonService
}

func NewHandler(service *payment_common_service.PaymentCommonService) *PaymentCommonHandler {
	return &PaymentCommonHandler{service: service}
}

func (h *PaymentCommonHandler) Routes(allAllowedMiddleware func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(allAllowedMiddleware)

	r.Get("/", h.GetPaymentHistories)
	r.Get("/{id}", h.GetPaymentHistoryById)
	r.Post("/{id}/cancel", h.CancelPayment)

	return r
}

// WebhookRoute returns the handler for webhook (no auth)
func (h *PaymentCommonHandler) WebhookRoute() http.HandlerFunc {
	return h.HandleWebhook
}

// GetPaymentHistories godoc
// @Summary Get payment histories
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
// @Router /api/payment [get]
func (h *PaymentCommonHandler) GetPaymentHistories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get profile from context (ID is the Profile ID)
	claims, _ := internal_middleware.GetProfileFromContext(ctx)
	if claims == nil {
		response.ValidationFailed(w, r, map[string]string{"authorization": "PROFILE_ID_REQUIRED"})
		return
	}

	// Parse query params
	var search *string
	if s := r.URL.Query().Get("search"); s != "" {
		search = &s
	}

	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		status = &s
	}

	sortBy := r.URL.Query().Get("sortBy")
	if sortBy == "" {
		sortBy = "created_at"
	}

	sortDir := r.URL.Query().Get("sortDir")
	if sortDir == "" {
		sortDir = "desc"
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	svcFilter := payment_common_service.GetPaymentHistoriesFilter{
		ProfileID: claims.ID.String(), // ID is the Profile ID
		Search:    search,
		Status:    status,
		SortBy:    sortBy,
		SortDir:   sortDir,
		Page:      page,
		Limit:     limit,
	}

	data, pag, err := h.service.GetPaymentHistories(ctx, svcFilter)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	// Build filter for response (Search is string, not *string)
	searchStr := ""
	if search != nil {
		searchStr = *search
	}
	reqFilter := &filter.ReqFilter{
		Search: searchStr,
		Page:   page,
		Limit:  limit,
		SortBy: sortBy,
		Sort:   sortDir,
	}

	response.LIST(w, r, "PAYMENT_HISTORIES_RETRIEVED", data, reqFilter, pag)
}

// GetPaymentHistoryById godoc
// @Summary Get payment history by ID
// @Tags Payment
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} response.Response{data=payment_common_service.PaymentHistoryResponse}
// @Router /api/payment/{id} [get]
func (h *PaymentCommonHandler) GetPaymentHistoryById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	if id == "" {
		response.ValidationFailed(w, r, map[string]string{"id": "REQUIRED"})
		return
	}

	// Get profile from context (ID is the Profile ID)
	claims, _ := internal_middleware.GetProfileFromContext(ctx)
	if claims == nil {
		response.ValidationFailed(w, r, map[string]string{"authorization": "PROFILE_ID_REQUIRED"})
		return
	}

	data, err := h.service.GetPaymentHistoryById(ctx, id, claims.ID.String())
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "PAYMENT_HISTORY_RETRIEVED", data)
}

// CancelPayment godoc
// @Summary Cancel a pending payment
// @Tags Payment
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} response.Response{data=payment_common_service.PaymentHistoryResponse}
// @Router /api/payment/{id}/cancel [post]
func (h *PaymentCommonHandler) CancelPayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	if id == "" {
		response.ValidationFailed(w, r, map[string]string{"id": "REQUIRED"})
		return
	}

	// Get profile from context (ID is the Profile ID)
	claims, _ := internal_middleware.GetProfileFromContext(ctx)
	if claims == nil {
		response.ValidationFailed(w, r, map[string]string{"authorization": "PROFILE_ID_REQUIRED"})
		return
	}

	data, err := h.service.CancelPayment(ctx, id, claims.ID.String())
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
