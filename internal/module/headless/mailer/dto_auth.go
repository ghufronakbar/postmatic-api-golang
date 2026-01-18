// internal/module/headless/mailer/dto_auth.go
package mailer

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
