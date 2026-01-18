// internal/module/payment/image_token/service/viewmodel.go
package image_token_service

import "time"

// ReferralInfo represents referral validation info in response
type ReferralInfo struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

// PriceCalculation represents the calculated price breakdown
type PriceCalculation struct {
	ItemPrice         int64 `json:"itemPrice"`         // harga asli token
	DiscountAmount    int64 `json:"discountAmount"`    // nominal diskon
	AfterDiscount     int64 `json:"afterDiscount"`     // harga setelah diskon
	AdminFeeAmount    int64 `json:"adminFeeAmount"`    // nominal admin fee
	SubtotalBeforeTax int64 `json:"subtotalBeforeTax"` // setelah diskon + admin
	TaxAmount         int64 `json:"taxAmount"`         // nominal tax
	TotalAmount       int64 `json:"totalAmount"`       // grand total
}

// PaymentMethodInfo represents payment method info in response
type PaymentMethodInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// CheckPriceResponse is the response for check price endpoint
// Note: Optional pointer fields will return null (not undefined) when nil
type CheckPriceResponse struct {
	Referral      *ReferralInfo     `json:"referral"`
	Calculation   PriceCalculation  `json:"calculation"`
	PaymentMethod PaymentMethodInfo `json:"paymentMethod"`
	TokenAmount   int64             `json:"tokenAmount"`
}

// CreatePaymentResponse is the response for create payment endpoint
// Note: Optional pointer fields will return null (not undefined) when nil
type CreatePaymentResponse struct {
	PaymentID     string            `json:"paymentId"`
	OrderID       string            `json:"orderId"`
	Status        string            `json:"status"`
	PaymentMethod PaymentMethodInfo `json:"paymentMethod"`
	ExpiresAt     *time.Time        `json:"expiresAt"`

	// Price calculation details
	Calculation PriceCalculation `json:"calculation"`
	TokenAmount int64            `json:"tokenAmount"`

	// Actions from Midtrans (filtered to public only)
	Actions []PaymentActionResponse `json:"actions"`
}

// PaymentActionResponse represents stored action from Midtrans
type PaymentActionResponse struct {
	Name      string `json:"name"`      // dari midtrans (generate-qr-code, deeplink-redirect)
	Label     string `json:"label"`     // label readable (QR Code, Virtual Account)
	Value     string `json:"value"`     // url atau va number
	ValueType string `json:"valueType"` // image, link, text, claim
	Method    string `json:"method"`    // GET, POST
}
