// internal/module/app/token_product/filter.go
package token_product

type TokenCalculateProductFilter struct {
	Type         string `json:"type" validate:"required,oneof=image_token video_token livestream_token"`
	Amount       int64  `json:"amount" validate:"required,gt=0"`
	CurrencyCode string `json:"currencyCode" validate:"required,len=3"`
	From         string `json:"from" validate:"required,oneof=price token"`
}
