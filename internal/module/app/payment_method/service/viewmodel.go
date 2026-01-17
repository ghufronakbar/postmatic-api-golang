package payment_method_service

import "time"

type PaymentMethodResponse struct {
	ID        int64                  `json:"id"`
	Code      string                 `json:"code"`
	Name      string                 `json:"name"`
	Type      PaymentMethodType      `json:"type"`
	Image     *string                `json:"image"`
	TaxFee    int64                  `json:"taxFee"`
	AdminType PaymentMethodAdminType `json:"adminType"`
	AdminFee  int64                  `json:"adminFee"`
	IsActive  bool                   `json:"isActive"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
}
