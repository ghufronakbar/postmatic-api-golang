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

// VERIFICATION EMAIL
type verificationInput struct {
	Name       string `json:"Name"`
	ConfirmUrl string `json:"ConfirmUrl"`
}

type VerificationInputDTO struct {
	Name  string `json:"Name"`
	To    string `json:"To" validate:"required,email"`
	Token string `json:"Token"`
	From  string `json:"From"`
}

// WELCOME EMAIL
type welcomeInput struct {
	Name  string `json:"Name"`
	Email string `json:"Email"`
	Link  string `json:"Link"`
}

type WelcomeInputDTO struct {
	Name  string `json:"Name"`
	Email string `json:"Email"`
	From  string `json:"From"`
}

// INVITATION EMAIL
type InvitationInputDTO struct {
	Email        string `json:"Email"`
	ConfirmUrl   string `json:"ConfirmUrl"`
	BusinessName string `json:"BusinessName"`
}
