// internal/module/headless/token/dto.go
package token

import "github.com/google/uuid"

type GenerateAccessTokenInput struct {
	ID       uuid.UUID
	Email    string
	Name     string
	ImageUrl *string
}

type GenerateRefreshTokenInput struct {
	ID    uuid.UUID
	Email string
}

type GenerateCreateAccountTokenInput struct {
	ID       uuid.UUID
	Email    string
	Name     string
	ImageUrl *string
}
