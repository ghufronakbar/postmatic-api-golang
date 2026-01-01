package session_repository

import (
	"time"

	"github.com/google/uuid"
)

type RedisSession struct {
	ID           uuid.UUID `json:"id"`
	RefreshToken string    `json:"refreshToken"`
	Browser      string    `json:"browser"`
	Platform     string    `json:"platform"` // OS
	Device       string    `json:"device"`   // Mobile/Desktop
	ClientIP     string    `json:"clientIp"` // IP Address
	ProfileID    uuid.UUID `json:"profileId"`
	CreatedAt    time.Time `json:"createdAt"`
	ExpiredAt    time.Time `json:"expiredAt"`
}
