// internal/module/app/social_platform/service/viewmodel.go
package social_platform_service

import "time"

// SocialPlatformResponse is response for social platform
type SocialPlatformResponse struct {
	ID           int64     `json:"id"`
	PlatformCode string    `json:"platformCode"`
	Logo         *string   `json:"logo"`
	Name         string    `json:"name"`
	Hint         string    `json:"hint"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// PlatformCodeResponse is response for platform code list
type PlatformCodeResponse struct {
	Codes []string `json:"codes"`
}
