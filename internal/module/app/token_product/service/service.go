// internal/module/app/token_product/service.go
package token_product_service

import (
	"context"
	"database/sql"
	"strings"

	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
)

type TokenProductService struct {
	store entity.Store
}

func NewTokenProductService(store entity.Store) *TokenProductService {
	return &TokenProductService{store: store}
}

// calculate token product based on price or token amount (dapat digunakan pada dashboard admin ataupun service untuk cek harga checkout sebelum tax/admin fee)
func (s *TokenProductService) CalculateTokenProduct(ctx context.Context, filter TokenCalculateProductFilter) (TokenCalculateProductResponse, error) {
	var res TokenCalculateProductResponse

	tokenType, err := validateTokenProductFilter(filter.Type)
	if err != nil {
		return res, err
	}

	if filter.From != "price" && filter.From != "token" {
		return res, errs.NewValidationFailed(map[string]string{"from": "INVALID_FROM"})
	}

	data, err := s.store.GetAppTokenProductByTypeCurrency(ctx, entity.GetAppTokenProductByTypeCurrencyParams{
		TokenType:    tokenType,
		CurrencyCode: strings.ToUpper(filter.CurrencyCode),
	})
	if err == sql.ErrNoRows {
		return res, errs.NewNotFound("TOKEN_PRODUCT_NOT_FOUND")
	}
	if err != nil {
		return res, err
	}

	if filter.From == "price" {
		res.TokenAmount, err = s.convertIDRToTokens(filter.Amount, data.PriceAmount, data.TokenAmount)
		if err != nil {
			return res, err
		}
		res.PriceAmount = filter.Amount
	} else {
		res.TokenAmount = filter.Amount
		res.PriceAmount, err = s.convertTokensToIDR(filter.Amount, data.PriceAmount, data.TokenAmount)
		if err != nil {
			return res, err
		}
	}

	res.CreatedAt = data.CreatedAt
	res.UpdatedAt = data.UpdatedAt
	res.ID = data.ID
	res.Type = string(data.TokenType)
	res.CurrencyCode = data.CurrencyCode

	return res, nil
}
