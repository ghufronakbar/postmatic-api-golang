// internal/module/headless/token/dto.go
package token

type GenerateAccessTokenInput struct {
	ID       string
	Email    string
	Name     string
	ImageUrl *string
}

type GenerateRefreshTokenInput struct {
	ID    string
	Email string
}

type GenerateCreateAccountTokenInput struct {
	ID       string
	Email    string
	Name     string
	ImageUrl *string
}
