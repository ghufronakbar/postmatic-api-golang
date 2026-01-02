// internal/module/headless/token/invitation_token.go
package token

import (
	"postmatic-api/pkg/errs"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type InvitationTokenClaims struct {
	// Profile ProfileID
	ProfileID             uuid.UUID `json:"profileId"`
	MemberID              int64     `json:"memberId"`
	BusinessRootID        int64     `json:"businessRootId"`
	MemberRole            string    `json:"memberRole"`
	MemberHistoryStatusID int64     `json:"memberHistoryStatusId"`
	jwt.RegisteredClaims
}

type GenerateInvitationTokenInput struct {
	// Profile ProfileID
	ProfileID             uuid.UUID
	MemberID              int64
	BusinessRootID        int64
	MemberRole            string
	MemberHistoryStatusID int64
}

func (tm *TokenMaker) GenerateInvitationToken(input GenerateInvitationTokenInput) (string, error) {
	expirationTime := time.Now().Add(tm.invitationTTL)
	claims := &InvitationTokenClaims{
		ProfileID:             input.ProfileID,
		MemberID:              input.MemberID,
		BusinessRootID:        input.BusinessRootID,
		MemberRole:            input.MemberRole,
		MemberHistoryStatusID: input.MemberHistoryStatusID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.invitationSecret)
}

func (tm *TokenMaker) ValidateInvitationToken(tokenString string) (*InvitationTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &InvitationTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return tm.invitationSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errs.NewBadRequest("INVALID_INVITATION_TOKEN")
	}
	return token.Claims.(*InvitationTokenClaims), nil
}

func (tm *TokenMaker) InvitationDecodeTokenWithoutVerify(tokenString string) (*InvitationTokenClaims, error) {
	parser := jwt.NewParser(
		jwt.WithoutClaimsValidation(),
	)

	claims := &InvitationTokenClaims{}
	_, _, err := parser.ParseUnverified(tokenString, claims)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
