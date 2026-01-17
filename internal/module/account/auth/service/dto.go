// internal/module/account/auth/dto.go
package auth_service

import (
	"postmatic-api/pkg/utils"
)

type LoginCredentialInput struct {
	From     string `json:"from" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterInput struct {
	Name     string `json:"name" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	From     string `json:"from" validate:"required"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type SessionInput struct {
	DeviceInfo utils.ClientInfo
}

type ResendEmailVerificationInput struct {
	From  string `json:"from" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type SubmitVerifyTokenInput struct {
	Token string `json:"token" validate:"required"`
	From  string `json:"from" validate:"required"`
}
