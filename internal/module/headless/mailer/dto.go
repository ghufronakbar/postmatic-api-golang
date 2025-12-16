// internal/module/headless/mailer/dto.go
package mailer

type SendEmailInput struct {
	To      string `json:"to"`
	Subject string `json:"subject"`

	// Opsi 1: Kirim Raw HTML/Text (Logic lama)
	Body string `json:"body"`
	Type string `json:"type"` // "html" or "text"

	// Opsi 2: Pakai Template (Logic baru)
	TemplateName EmailTemplate `json:"templateName"`
	Data         interface{}   `json:"data"` // data untuk di-inject ke template
}

type VerificationInput struct {
	Name       string `json:"Name"`
	ConfirmUrl string `json:"ConfirmUrl"`
}

type WelcomeInput struct {
	Name  string `json:"Name"`
	Email string `json:"Email"`
	Link  string `json:"Link"`
}
