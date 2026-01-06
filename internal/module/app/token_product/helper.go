// internal/module/app/token_product/helper.go
package token_product

import (
	"errors"
	"math"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
)

func mulWillOverflow(a, b int64) bool {
	if a == 0 || b == 0 {
		return false
	}
	return a > math.MaxInt64/b
}

// convert amountIDR to tokens (token dibulatkan kebawah)
func (s *TokenProductService) convertIDRToTokens(amountIDR, priceBase, tokenBase int64) (int64, error) {
	// asumsi semua > 0 sudah divalidasi sebelumnya
	if mulWillOverflow(amountIDR, tokenBase) {
		return 0, errs.NewInternalServerError(errors.New("OVERFLOW_CONVERT_IDR_TO_TOKEN"))
	}
	return (amountIDR * tokenBase) / priceBase, nil
}

// convert tokens to amountIDR (amountIDR dibulatkan keatas)
func (s *TokenProductService) convertTokensToIDR(tokens, priceBase, tokenBase int64) (int64, error) {
	if mulWillOverflow(tokens, priceBase) {
		return 0, errs.NewInternalServerError(errors.New("OVERFLOW_CONVERT_TOKEN_TO_IDR"))
	}
	num := tokens * priceBase
	quo := num / tokenBase
	if num%tokenBase > 0 {
		return quo + 1, nil
	}
	return quo, nil
}

// validate token type
func validateTokenProductFilter(typeToken string) (entity.TokenType, error) {
	if typeToken != "image_token" && typeToken != "video_token" && typeToken != "livestream_token" {
		return "", errs.NewBadRequest("TOKEN_TYPE_NOT_FOUND")
	}
	return entity.TokenType(typeToken), nil
}
