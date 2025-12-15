// internal/module/account/auth/dto.go
package auth

import (
	"postmatic-api/pkg/utils"
)

type LoginCredentialsInput struct {
	Locale   string `json:"locale" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterInput struct {
	Name     string `json:"name" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Locale   string `json:"locale" validate:"required"`
}

type RefreshTokenInput struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type SessionInput struct {
	DeviceInfo utils.ClientInfo
}

type ResendEmailVerificationInput struct {
	Locale string `json:"locale" validate:"required"`
	Email  string `json:"email" validate:"required,email"`
}
