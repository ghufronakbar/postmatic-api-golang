package session

type LogoutInput struct {
	SessionID string `json:"sessionId" validate:"required"`
}

type LogoutAllInput struct {
	ProfileID string `json:"profileId" validate:"required"`
}
