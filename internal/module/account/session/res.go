package session

import (
	"postmatic-api/internal/repository/redis"
	"time"
)

type SessionResponse struct {
	ID       string             `json:"id"`
	Email    string             `json:"email"`
	Name     string             `json:"name"`
	ImageUrl *string            `json:"imageUrl"`
	Exp      int64              `json:"exp"`
	Session  redis.RedisSession `json:"session"`
}

type SessionListResponse struct {
	ID        string    `json:"id"`
	Browser   string    `json:"browser"`
	Platform  string    `json:"platform"` // OS
	Device    string    `json:"device"`   // Mobile/Desktop
	ClientIP  string    `json:"clientIp"` // IP Address
	ProfileID string    `json:"profileId"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiredAt time.Time `json:"expiredAt"`
}
