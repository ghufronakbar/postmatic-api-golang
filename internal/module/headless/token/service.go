// internal/module/headless/token/service.go
package token

import (
	"errors"
	"postmatic-api/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	// Profile ID
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	ImageUrl *string   `json:"imageUrl"`
	jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	jwt.RegisteredClaims
}

type TokenMaker struct {
	accessSecret        []byte
	refreshSecret       []byte
	createAccountSecret []byte
	accessTTL           time.Duration
	refreshTTL          time.Duration
	createAccountTTL    time.Duration
}

func NewTokenMaker(cfg *config.Config) *TokenMaker {
	return &TokenMaker{
		accessSecret:        []byte(cfg.JWT_ACCESS_TOKEN_SECRET),
		refreshSecret:       []byte(cfg.JWT_REFRESH_TOKEN_SECRET),
		createAccountSecret: []byte(cfg.JWT_CREATE_ACCOUNT_TOKEN_SECRET),
		accessTTL:           cfg.JWT_ACCESS_TOKEN_EXPIRED,
		refreshTTL:          cfg.JWT_REFRESH_TOKEN_EXPIRED,
		createAccountTTL:    cfg.JWT_CREATE_ACCOUNT_TOKEN_EXPIRED,
	}
}

// JWT AUTH
func (tm *TokenMaker) GenerateAccessToken(input GenerateAccessTokenInput) (string, error) {
	expirationTime := time.Now().Add(tm.accessTTL)
	claims := &Claims{
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

func (tm *TokenMaker) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return tm.accessSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("INVALID_ACCESS_TOKEN")
	}
	return token.Claims.(*Claims), nil
}

func (tm *TokenMaker) ValidateRefreshToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return tm.refreshSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("INVALID_REFRESH_TOKEN")
	}
	return token.Claims.(*Claims), nil
}

// JWT CREATE ACCOUNT

func (tm *TokenMaker) GenerateCreateAccountToken(input GenerateCreateAccountTokenInput) (string, error) {
	expirationTime := time.Now().Add(tm.createAccountTTL)
	claims := &Claims{
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

func (tm *TokenMaker) ValidateCreateAccountToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return tm.createAccountSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("INVALID_CREATE_ACCOUNT_TOKEN")
	}
	return token.Claims.(*Claims), nil
}

// DECODE TOKEN
func (tm *TokenMaker) DecodeTokenWithoutVerify(tokenString string) (*Claims, error) {
	parser := jwt.NewParser(
		jwt.WithoutClaimsValidation(),
	)

	claims := &Claims{}
	_, _, err := parser.ParseUnverified(tokenString, claims)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
