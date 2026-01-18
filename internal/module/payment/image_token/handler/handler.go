// internal/module/payment/image_token/handler/handler.go
package image_token_handler

import (
	"net/http"
	"strconv"

	"postmatic-api/internal/internal_middleware"
	image_token_service "postmatic-api/internal/module/payment/image_token/service"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type ImageTokenPaymentHandler struct {
	service *image_token_service.ImageTokenPaymentService
}

func NewHandler(service *image_token_service.ImageTokenPaymentService) *ImageTokenPaymentHandler {
	return &ImageTokenPaymentHandler{service: service}
}

func (h *ImageTokenPaymentHandler) Routes(allAllowedMiddleware func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(allAllowedMiddleware)

	r.Get("/", h.CheckPrice)
	r.Post("/", h.CreatePayment)

	return r
}

// CheckPrice godoc
// @Summary Check price for image token purchase
// @Tags Payment
// @Accept json
// @Produce json
// @Param tokenAmount query int true "Token amount to purchase"
// @Param currencyCode query string true "Currency code (e.g., IDR)"
// @Param paymentMethod query string true "Payment method code (e.g., bca, gopay)"
// @Param referralCode query string false "Referral code (optional)"
// @Param businessRootId query int true "Business root ID"
// @Success 200 {object} response.Response{data=image_token_service.CheckPriceResponse}
// @Router /api/payment/image-token [get]
func (h *ImageTokenPaymentHandler) CheckPrice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query params
	tokenAmountStr := r.URL.Query().Get("tokenAmount")
	if tokenAmountStr == "" {
		response.ValidationFailed(w, r, map[string]string{"tokenAmount": "REQUIRED"})
		return
	}
	tokenAmount, err := strconv.ParseInt(tokenAmountStr, 10, 64)
	if err != nil || tokenAmount <= 0 {
		response.ValidationFailed(w, r, map[string]string{"tokenAmount": "MUST_BE_POSITIVE_INTEGER"})
		return
	}

	currencyCode := r.URL.Query().Get("currencyCode")
	if currencyCode == "" {
		response.ValidationFailed(w, r, map[string]string{"currencyCode": "REQUIRED"})
		return
	}

	paymentMethod := r.URL.Query().Get("paymentMethod")
	if paymentMethod == "" {
		response.ValidationFailed(w, r, map[string]string{"paymentMethod": "REQUIRED"})
		return
	}

	businessRootIdStr := r.URL.Query().Get("businessRootId")
	if businessRootIdStr == "" {
		response.ValidationFailed(w, r, map[string]string{"businessRootId": "REQUIRED"})
		return
	}
	businessRootId, err := strconv.ParseInt(businessRootIdStr, 10, 64)
	if err != nil {
		response.ValidationFailed(w, r, map[string]string{"businessRootId": "MUST_BE_INTEGER"})
		return
	}

	var referralCode *string
	if rc := r.URL.Query().Get("referralCode"); rc != "" {
		referralCode = &rc
	}

	// Get profile from context (ID is the Profile ID)
	claims, _ := internal_middleware.GetProfileFromContext(ctx)
	if claims == nil {
		response.ValidationFailed(w, r, map[string]string{"authorization": "PROFILE_ID_REQUIRED"})
		return
	}

	input := image_token_service.CheckPriceInput{
		TokenAmount:    tokenAmount,
		CurrencyCode:   currencyCode,
		PaymentMethod:  paymentMethod,
		ReferralCode:   referralCode,
		BusinessRootID: businessRootId,
		ProfileID:      claims.ID, // ID is the Profile ID
	}

	result, err := h.service.CheckPrice(ctx, input)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "CHECK_PRICE_SUCCESS", result)
}

// CreatePayment godoc
// @Summary Create payment for image token purchase
// @Tags Payment
// @Accept json
// @Produce json
// @Param body body image_token_service.CreatePaymentInput true "Payment input"
// @Success 201 {object} response.Response{data=image_token_service.CreatePaymentResponse}
// @Router /api/payment/image-token [post]
func (h *ImageTokenPaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var input image_token_service.CreatePaymentInput

	// Validate request body with utils.ValidateStruct
	if appErr := utils.ValidateStruct(r.Body, &input); appErr != nil {
		response.ValidationFailed(w, r, appErr.ValidationErrors)
		return
	}

	// Get profile from context (ID is the Profile ID)
	claims, _ := internal_middleware.GetProfileFromContext(ctx)
	if claims == nil {
		response.ValidationFailed(w, r, map[string]string{"authorization": "PROFILE_ID_REQUIRED"})
		return
	}
	input.ProfileID = claims.ID // ID is the Profile ID

	result, err := h.service.CreatePayment(ctx, input)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "PAYMENT_CREATED", result)
}
