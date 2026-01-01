// internal/module/headless/token/create_account_token.go
package token

import (
	"postmatic-api/pkg/errs"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type CreateAccountTokenClaims struct {
	// Profile ID
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	ImageUrl *string   `json:"imageUrl"`
	jwt.RegisteredClaims
}

type GenerateCreateAccountTokenInput struct {
	ID       uuid.UUID
	Email    string
	Name     string
	ImageUrl *string
}

func (tm *TokenMaker) GenerateCreateAccountToken(input GenerateCreateAccountTokenInput) (string, error) {
	expirationTime := time.Now().Add(tm.createAccountTTL)
	claims := &CreateAccountTokenClaims{
		ID:       input.ID,
		Email:    input.Email,
		Name:     input.Name,
		ImageUrl: input.ImageUrl,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.createAccountSecret)
}

func (tm *TokenMaker) ValidateCreateAccountToken(tokenString string) (*CreateAccountTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CreateAccountTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return tm.createAccountSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errs.NewBadRequest("INVALID_CREATE_ACCOUNT_TOKEN")
	}
	return token.Claims.(*CreateAccountTokenClaims), nil
}
