// internal/module/headless/token/service.go
package token

import (
	"postmatic-api/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenMaker struct {
	// ACCESS
	accessSecret []byte
	accessTTL    time.Duration
	// REFRESH
	refreshSecret []byte
	refreshTTL    time.Duration
	// CREATE ACCOUNT
	createAccountSecret []byte
	createAccountTTL    time.Duration
	// INVITATION
	invitationSecret []byte
	invitationTTL    time.Duration
}

func NewTokenMaker(cfg *config.Config) *TokenMaker {
	return &TokenMaker{
		accessSecret:        []byte(cfg.JWT_ACCESS_TOKEN_SECRET),
		refreshSecret:       []byte(cfg.JWT_REFRESH_TOKEN_SECRET),
		createAccountSecret: []byte(cfg.JWT_CREATE_ACCOUNT_TOKEN_SECRET),
		accessTTL:           cfg.JWT_ACCESS_TOKEN_EXPIRED,
		refreshTTL:          cfg.JWT_REFRESH_TOKEN_EXPIRED,
		createAccountTTL:    cfg.JWT_CREATE_ACCOUNT_TOKEN_EXPIRED,
		invitationSecret:    []byte(cfg.JWT_INVITATION_TOKEN_SECRET),
		invitationTTL:       cfg.JWT_INVITATION_TOKEN_EXPIRED,
	}
}

// DECODE TOKEN
func (tm *TokenMaker) DecodeTokenWithoutVerify(tokenString string) (*AccessTokenClaims, error) {
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
