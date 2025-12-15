// pkg/token/token.go
package token

import (
	"errors"
	"postmatic-api/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	// Profile ID
	ID       string  `json:"id"`
	Email    string  `json:"email"`
	Name     string  `json:"name"`
	ImageUrl *string `json:"imageUrl"`
	jwt.RegisteredClaims
}

// JWT AUTH
func GenerateAccessToken(ID, email, name string, imageUrl *string) (string, error) {
	expirationTime := time.Now().Add(config.Load().JWT_ACCESS_TOKEN_EXPIRED)
	claims := &Claims{
		ID:       ID,
		Email:    email,
		Name:     name,
		ImageUrl: imageUrl,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Load().JWT_ACCESS_TOKEN_SECRET))
}

func GenerateRefreshToken(ID, email, name string, imageUrl *string) (string, error) {
	expirationTime := time.Now().Add(config.Load().JWT_REFRESH_TOKEN_EXPIRED)
	claims := &Claims{
		ID:       ID,
		Email:    email,
		Name:     name,
		ImageUrl: imageUrl,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Load().JWT_REFRESH_TOKEN_SECRET))
}

func ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Load().JWT_ACCESS_TOKEN_SECRET), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("INVALID_ACCESS_TOKEN")
	}
	return token.Claims.(*Claims), nil
}

func ValidateRefreshToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Load().JWT_REFRESH_TOKEN_SECRET), nil
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

func GenerateCreateAccountToken(ID, email, name string, imageUrl *string) (string, error) {
	expirationTime := time.Now().Add(config.Load().JWT_CREATE_ACCOUNT_TOKEN_EXPIRED)
	claims := &Claims{
		ID:       ID,
		Email:    email,
		Name:     name,
		ImageUrl: imageUrl,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Load().JWT_CREATE_ACCOUNT_TOKEN_SECRET))
}

func ValidateCreateAccountToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Load().JWT_CREATE_ACCOUNT_TOKEN_SECRET), nil
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
func DecodeTokenWithoutVerify(tokenString string) (*Claims, error) {
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
