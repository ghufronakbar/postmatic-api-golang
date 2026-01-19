// internal/module/app/social_platform/service/dto.go
package social_platform_service

import "github.com/google/uuid"

// CreateSocialPlatformInput is input for creating social platform
type CreateSocialPlatformInput struct {
	PlatformCode string  `json:"platformCode" validate:"required"`
	Logo         *string `json:"logo"`
	Name         string  `json:"name" validate:"required"`
	Hint         string  `json:"hint" validate:"required"`
	IsActive     bool    `json:"isActive"`
	ProfileID    uuid.UUID
}

// UpdateSocialPlatformInput is input for updating social platform
type UpdateSocialPlatformInput struct {
	PlatformCode string  `json:"platformCode" validate:"required"`
	Logo         *string `json:"logo"`
	Name         string  `json:"name" validate:"required"`
	Hint         string  `json:"hint" validate:"required"`
	IsActive     bool    `json:"isActive"`
	ProfileID    uuid.UUID
}

// GetSocialPlatformsFilter is input for filtering social platforms
type GetSocialPlatformsFilter struct {
	IncludeInactive bool
	Search          *string
	SortBy          string
	SortDir         string
	Page            int
	Limit           int
}
