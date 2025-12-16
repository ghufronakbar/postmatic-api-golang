package session_repository

import "time"

type RedisSession struct {
	ID           string    `json:"id"`
	RefreshToken string    `json:"refreshToken"`
	Browser      string    `json:"browser"`
	Platform     string    `json:"platform"` // OS
	Device       string    `json:"device"`   // Mobile/Desktop
	ClientIP     string    `json:"clientIp"` // IP Address
	ProfileID    string    `json:"profileId"`
	CreatedAt    time.Time `json:"createdAt"`
	ExpiredAt    time.Time `json:"expiredAt"`
}
