// internal/http/handler/app_handler/token_product_handler.go
package app_handler

import (
	"net/http"
	"postmatic-api/internal/module/app/token_product"
	"postmatic-api/internal/repository/entity"

	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"

	"github.com/go-chi/chi/v5"
)

type TokenProductHandler struct {
	svc *token_product.TokenProductService
}

func NewTokenProductHandler(svc *token_product.TokenProductService) *TokenProductHandler {
	return &TokenProductHandler{svc: svc}
}

func (h *TokenProductHandler) TokenProductRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.CalculateTokenProduct)

	return r
}

func (h *TokenProductHandler) CalculateTokenProduct(w http.ResponseWriter, r *http.Request) {
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
	filter := token_product.TokenCalculateProductFilter{
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
