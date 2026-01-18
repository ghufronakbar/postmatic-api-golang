// internal/module/payment/common/service/viewmodel.go
package payment_common_service

import "time"

// PaymentHistoryResponse is the response for payment history endpoints
// Note: Optional pointer fields will return null (not undefined) when nil
type PaymentHistoryResponse struct {
	ID                 string     `json:"id"`
	ProductAmount      int64      `json:"productAmount"`
	Status             string     `json:"status"`
	Currency           string     `json:"currency"`
	PaymentMethod      string     `json:"paymentMethod"`
	PaymentMethodType  string     `json:"paymentMethodType"`
	ProductName        string     `json:"productName"`
	ProductType        string     `json:"productType"`
	ProductPrice       int64      `json:"productPrice"`
	ProductImageUrl    string     `json:"productImageUrl"`
	SubtotalItemAmount int64      `json:"subtotalItemAmount"`
	DiscountAmount     int64      `json:"discountAmount"`
	AdminFeeAmount     int64      `json:"adminFeeAmount"`
	TaxAmount          int64      `json:"taxAmount"`
	TotalAmount        int64      `json:"totalAmount"`
	MidtransExpiredAt  *time.Time `json:"midtransExpiredAt"`
	PaymentPendingAt   *time.Time `json:"paymentPendingAt"`
	PaymentSuccessAt   *time.Time `json:"paymentSuccessAt"`
	PaymentFailedAt    *time.Time `json:"paymentFailedAt"`
	PaymentCanceledAt  *time.Time `json:"paymentCanceledAt"`
	PaymentExpiredAt   *time.Time `json:"paymentExpiredAt"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}
