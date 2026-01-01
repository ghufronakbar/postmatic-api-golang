// internal/module/headless/token/refresh_token.go
package token

import (
	"postmatic-api/pkg/errs"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RefreshTokenClaims struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	jwt.RegisteredClaims
}

type GenerateRefreshTokenInput struct {
	ID    uuid.UUID
	Email string
}

func (tm *TokenMaker) GenerateRefreshToken(input GenerateRefreshTokenInput) (string, error) {
	expirationTime := time.Now().Add(tm.refreshTTL)
	claims := &RefreshTokenClaims{
		ID:    input.ID,
		Email: input.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.refreshSecret)
}

func (tm *TokenMaker) ValidateRefreshToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return tm.refreshSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errs.NewBadRequest("INVALID_REFRESH_TOKEN")
	}
	return token.Claims.(*AccessTokenClaims), nil
}
