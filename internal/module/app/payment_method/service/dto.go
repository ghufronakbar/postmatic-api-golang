package payment_method_service

type CreatePaymentMethodInput struct {
	ProfileID string                 `json:"-"`
	Code      string                 `json:"code" validate:"required,max=20"`
	Name      string                 `json:"name" validate:"required,max=255"`
	Type      PaymentMethodType      `json:"type" validate:"required,oneof=bank ewallet"`
	Image     *string                `json:"image" validate:"omitempty,url,max=255"`
	TaxFee    int64                  `json:"taxFee" validate:"min=0,max=100"`
	AdminType PaymentMethodAdminType `json:"adminType" validate:"required,oneof=fixed percentage"`
	AdminFee  int64                  `json:"adminFee" validate:"min=0"`
	IsActive  bool                   `json:"isActive"`
}

type UpdatePaymentMethodInput struct {
	ID        int64                  `json:"-"`
	ProfileID string                 `json:"-"`
	Code      string                 `json:"code" validate:"required,max=20"`
	Name      string                 `json:"name" validate:"required,max=255"`
	Type      PaymentMethodType      `json:"type" validate:"required,oneof=bank ewallet"`
	Image     *string                `json:"image" validate:"omitempty,url,max=255"`
	TaxFee    int64                  `json:"taxFee" validate:"min=0,max=100"`
	AdminType PaymentMethodAdminType `json:"adminType" validate:"required,oneof=fixed percentage"`
	AdminFee  int64                  `json:"adminFee" validate:"min=0"`
	IsActive  bool                   `json:"isActive"`
}

type PaymentMethodType string

const (
	PaymentMethodTypeBank    PaymentMethodType = "bank"
	PaymentMethodTypeEwallet PaymentMethodType = "ewallet"
)

type PaymentMethodAdminType string

const (
	PaymentMethodAdminTypeFixed      PaymentMethodAdminType = "fixed"
	PaymentMethodAdminTypePercentage PaymentMethodAdminType = "percentage"
)
