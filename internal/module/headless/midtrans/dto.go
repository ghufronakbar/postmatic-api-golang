package midtrans

// ================== INPUT DTOs ==================

// CustomerDetails represents customer information for payment
type CustomerDetails struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

// ItemDetail represents an item in the transaction
type ItemDetail struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Price    int64  `json:"price"`
	Quantity int32  `json:"quantity"`
}

// ChargeGopayInput is the input DTO for charging Gopay payments
type ChargeGopayInput struct {
	OrderID         string          `json:"orderId" validate:"required"`
	GrossAmount     int64           `json:"grossAmount" validate:"required,min=1"`
	CustomerDetails CustomerDetails `json:"customerDetails"`
	Items           []ItemDetail    `json:"items"`
	// Optional: callback URL for Gopay redirect
	CallbackURL string `json:"callbackUrl"`
}

// ChargeBankTransferInput is the input DTO for charging Bank Transfer payments
type ChargeBankTransferInput struct {
	OrderID         string          `json:"orderId" validate:"required"`
	GrossAmount     int64           `json:"grossAmount" validate:"required,min=1"`
	Bank            string          `json:"bank" validate:"required,oneof=bca bni bri permata mandiri"` // Bank name
	CustomerDetails CustomerDetails `json:"customerDetails"`
	Items           []ItemDetail    `json:"items"`
}

// ================== OUTPUT DTOs ==================

// ChargeResponse is the output DTO for charge operations
type ChargeResponse struct {
	TransactionID     string `json:"transactionId"`
	OrderID           string `json:"orderId"`
	GrossAmount       string `json:"grossAmount"`
	PaymentType       string `json:"paymentType"`
	TransactionTime   string `json:"transactionTime"`
	TransactionStatus string `json:"transactionStatus"`
	FraudStatus       string `json:"fraudStatus"`
	StatusCode        string `json:"statusCode"`
	StatusMessage     string `json:"statusMessage"`
	// E-Wallet specific
	Actions []PaymentAction `json:"actions,omitempty"`
	// Bank Transfer specific (VA Numbers)
	VANumbers       []VANumber `json:"vaNumbers,omitempty"`
	PermataVANumber string     `json:"permataVaNumber,omitempty"`
}

// PaymentAction represents action URLs for e-wallet payments
type PaymentAction struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	URL    string `json:"url"`
}

// VANumber represents Virtual Account number for bank transfer
type VANumber struct {
	Bank     string `json:"bank"`
	VANumber string `json:"vaNumber"`
}

// TransactionStatusResponse is the output DTO for transaction status operations
type TransactionStatusResponse struct {
	TransactionID     string `json:"transactionId"`
	OrderID           string `json:"orderId"`
	GrossAmount       string `json:"grossAmount"`
	PaymentType       string `json:"paymentType"`
	TransactionTime   string `json:"transactionTime"`
	TransactionStatus string `json:"transactionStatus"`
	FraudStatus       string `json:"fraudStatus"`
	StatusCode        string `json:"statusCode"`
	StatusMessage     string `json:"statusMessage"`
}
