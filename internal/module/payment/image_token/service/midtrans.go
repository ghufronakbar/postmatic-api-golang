// internal/module/payment/image_token/service/midtrans.go
package image_token_service

// PaymentAction for e-wallet deep link (from Midtrans response)
type PaymentAction struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	URL    string `json:"url"`
}

// VANumber for bank transfer (from Midtrans response)
type VANumber struct {
	Bank     string `json:"bank"`
	VANumber string `json:"vaNumber"`
}
