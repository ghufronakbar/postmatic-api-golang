// internal/module/app/token_product/handler/handler.go
package token_product_handler

import (
	"net/http"
	token_product_service "postmatic-api/internal/module/app/token_product/service"
	"postmatic-api/internal/repository/entity"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *token_product_service.TokenProductService
}

func NewHandler(svc *token_product_service.TokenProductService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.CalculateTokenProduct)

	return r
}

func (h *Handler) CalculateTokenProduct(w http.ResponseWriter, r *http.Request) {
	amount, err := utils.GetQueryInt64(r, "amount")
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}
	currencyCode, err := utils.GetQueryEnum(r, "currencyCode", []string{"IDR"})
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}
	from, err := utils.GetQueryEnum(r, "from", []string{"price", "token"})
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}
	typeToken, err := utils.GetQueryEnum(r, "type", []string{
		string(entity.TokenTypeImageToken),
		string(entity.TokenTypeVideoToken),
		string(entity.TokenTypeLivestreamToken),
	})
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}
	filter := token_product_service.TokenCalculateProductFilter{
		Type:         typeToken,
		Amount:       amount,
		CurrencyCode: currencyCode,
		From:         from,
	}

	res, err := h.svc.CalculateTokenProduct(r.Context(), filter)
	if err != nil {
		response.Error(w, r, err, nil)
		return
	}

	response.OK(w, r, "SUCCESS_CALCULATE_TOKEN_PRODUCT", res)
}
