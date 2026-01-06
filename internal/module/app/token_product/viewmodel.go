// internal/module/app/token_product/viewmodel.go
package token_product

import (
	"time"

	"github.com/google/uuid"
)

type TokenCalculateProductResponse struct {
	ID           uuid.UUID `json:"id"`
	Type         string    `json:"type"`
	CurrencyCode string    `json:"currencyCode"`
	PriceAmount  int64     `json:"priceAmount"`
	TokenAmount  int64     `json:"tokenAmount"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
