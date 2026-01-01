// internal/module/account/session/viewmodel.go
package session

import (
	sessRepo "postmatic-api/internal/repository/redis/session_repository"
	"time"

	"github.com/google/uuid"
)

type SessionResponse struct {
	ID       uuid.UUID             `json:"id"`
	Email    string                `json:"email"`
	Name     string                `json:"name"`
	ImageUrl *string               `json:"imageUrl"`
	Exp      int64                 `json:"exp"`
	Session  sessRepo.RedisSession `json:"session"`
}

type SessionListResponse struct {
	ID        uuid.UUID `json:"id"`
	Browser   string    `json:"browser"`
	Platform  string    `json:"platform"` // OS
	Device    string    `json:"device"`   // Mobile/Desktop
	ClientIP  string    `json:"clientIp"` // IP Address
	ProfileID uuid.UUID `json:"profileId"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiredAt time.Time `json:"expiredAt"`
}
