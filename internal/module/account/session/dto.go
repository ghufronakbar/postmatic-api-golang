package session

import "github.com/google/uuid"

type LogoutInput struct {
	SessionID uuid.UUID `json:"sessionId" validate:"required"`
}

type LogoutAllInput struct {
	ProfileID uuid.UUID `json:"profileId" validate:"required"`
}
