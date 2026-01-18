// internal/module/headless/mailer/dto_payment.go
package mailer

import "time"

// PAYMENT CHECKOUT EMAIL
// Sent when user creates a new payment
type paymentCheckoutInput struct {
	Name          string               `json:"Name"`
	OrderID       string               `json:"OrderID"`
	ProductName   string               `json:"ProductName"`
	PaymentMethod string               `json:"PaymentMethod"`
	TotalAmount   string               `json:"TotalAmount"` // formatted dengan currency
	ExpiresAt     string               `json:"ExpiresAt"`   // formatted datetime
	Actions       []PaymentActionInput `json:"Actions"`
}

type PaymentActionInput struct {
	Label     string `json:"Label"`
	Value     string `json:"Value"`
	ValueType string `json:"ValueType"` // text, link, image
}

type PaymentCheckoutInputDTO struct {
	// recipient
	Email string `json:"Email"`
	Name  string `json:"Name"`

	// order info
	OrderID       string `json:"OrderID"`
	ProductName   string `json:"ProductName"`
	PaymentMethod string `json:"PaymentMethod"`
	TotalAmount   string `json:"TotalAmount"` // formatted dengan currency
	ExpiresAt     string `json:"ExpiresAt"`   // formatted datetime

	// payment actions (VA number, QR code link, dll)
	Actions []PaymentActionInput `json:"Actions"`
}

// PAYMENT SUCCESS EMAIL (Invoice)
// Sent when payment status changes to success
type paymentSuccessInput struct {
	Name        string `json:"Name"`
	OrderID     string `json:"OrderID"`
	ProductName string `json:"ProductName"`

	// calculation breakdown
	ItemPrice      string `json:"ItemPrice"`
	DiscountAmount string `json:"DiscountAmount"`
	AdminFeeAmount string `json:"AdminFeeAmount"`
	TaxAmount      string `json:"TaxAmount"`
	TotalAmount    string `json:"TotalAmount"`

	PaymentMethod string `json:"PaymentMethod"`
	PaidAt        string `json:"PaidAt"` // formatted datetime
}

type PaymentSuccessInputDTO struct {
	// recipient
	Email string `json:"Email"`
	Name  string `json:"Name"`

	// order info
	OrderID     string `json:"OrderID"`
	ProductName string `json:"ProductName"`

	// calculation breakdown
	ItemPrice      int64  `json:"ItemPrice"`
	DiscountAmount int64  `json:"DiscountAmount"`
	AdminFeeAmount int64  `json:"AdminFeeAmount"`
	TaxAmount      int64  `json:"TaxAmount"`
	TotalAmount    int64  `json:"TotalAmount"`
	Currency       string `json:"Currency"`

	PaymentMethod string    `json:"PaymentMethod"`
	PaidAt        time.Time `json:"PaidAt"`
}

// PAYMENT CANCELED EMAIL
// Sent when user cancels a pending payment
type paymentCanceledInput struct {
	Name          string `json:"Name"`
	OrderID       string `json:"OrderID"`
	ProductName   string `json:"ProductName"`
	TotalAmount   string `json:"TotalAmount"`
	PaymentMethod string `json:"PaymentMethod"`
	CanceledAt    string `json:"CanceledAt"` // formatted datetime
}

type PaymentCanceledInputDTO struct {
	// recipient
	Email string `json:"Email"`
	Name  string `json:"Name"`

	// order info
	OrderID       string    `json:"OrderID"`
	ProductName   string    `json:"ProductName"`
	TotalAmount   int64     `json:"TotalAmount"`
	Currency      string    `json:"Currency"`
	PaymentMethod string    `json:"PaymentMethod"`
	CanceledAt    time.Time `json:"CanceledAt"`
}
