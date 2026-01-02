// internal/module/headless/token/access_token.go
package token

import (
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type GenerateAccessTokenInput struct {
	ID       uuid.UUID
	Email    string
	Name     string
	ImageUrl *string
	Role     entity.AppRole
}

type AccessTokenClaims struct {
	// Profile ID
	ID       uuid.UUID      `json:"id"`
	Email    string         `json:"email"`
	Name     string         `json:"name"`
	ImageUrl *string        `json:"imageUrl"`
	Role     entity.AppRole `json:"role"`
	jwt.RegisteredClaims
}

func (tm *TokenMaker) GenerateAccessToken(input GenerateAccessTokenInput) (string, error) {
	expirationTime := time.Now().Add(tm.accessTTL)
	claims := &AccessTokenClaims{
		ID:       input.ID,
		Email:    input.Email,
		Name:     input.Name,
		ImageUrl: input.ImageUrl,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.accessSecret)
}

func (tm *TokenMaker) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return tm.accessSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errs.NewBadRequest("INVALID_ACCESS_TOKEN")
	}
	return token.Claims.(*AccessTokenClaims), nil
}

func (tm *TokenMaker) AccessDecodeTokenWithoutVerify(tokenString string) (*AccessTokenClaims, error) {
	parser := jwt.NewParser(
		jwt.WithoutClaimsValidation(),
	)

	claims := &AccessTokenClaims{}
	_, _, err := parser.ParseUnverified(tokenString, claims)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
