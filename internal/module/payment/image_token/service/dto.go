// internal/module/payment/image_token/service/dto.go
package image_token_service

import "github.com/google/uuid"

// CheckPriceInput is the input for checking price before payment
type CheckPriceInput struct {
	TokenAmount    int64   `json:"tokenAmount" validate:"required,min=1"`
	CurrencyCode   string  `json:"currencyCode" validate:"required"`
	PaymentMethod  string  `json:"paymentMethod" validate:"required"`
	ReferralCode   *string `json:"referralCode"`
	BusinessRootID int64   `json:"businessRootId" validate:"required"`
	ProfileID      uuid.UUID
}

// CreatePaymentInput is the input for creating a payment
type CreatePaymentInput struct {
	TokenAmount    int64   `json:"tokenAmount" validate:"required,min=1"`
	CurrencyCode   string  `json:"currencyCode" validate:"required"`
	PaymentMethod  string  `json:"paymentMethod" validate:"required"`
	ReferralCode   *string `json:"referralCode"`
	BusinessRootID int64   `json:"businessRootId" validate:"required"`
	ProfileID      uuid.UUID
}

// PriceCalculationInput is internal input for price calculation
type PriceCalculationInput struct {
	BasePrice     int64  // harga asli product
	DiscountType  string // "fixed" atau "percentage"
	DiscountValue int64  // nilai diskon (nominal atau %)
	MaxDiscount   int64  // max cap untuk percentage
	AdminFeeType  string // "fixed" atau "percentage"
	AdminFeeValue int64  // nilai admin fee
	TaxPercentage int64  // tax percentage
}
