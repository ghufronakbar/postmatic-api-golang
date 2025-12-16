// internal/module/account/profile/viewmodel.go
package profile

import "time"

type GetProfileResponse struct {
	// Profile ID
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Email       string         `json:"email"`
	ImageUrl    *string        `json:"imageUrl"`
	CountryCode string         `json:"countryCode"`
	Phone       string         `json:"phone"`
	Description *string        `json:"description"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	Credential  CredentialUser `json:"credential"`
	Google      GoogleUser     `json:"google"`
}

type CredentialUser struct {
	IsPasswordSet bool       `json:"isPasswordSet"`
	VerifiedAt    *time.Time `json:"verifiedAt"`
}

type GoogleUser struct {
	IsConnected bool       `json:"isConnected"`
	VerifiedAt  *time.Time `json:"verifiedAt"`
}

type UpdateProfileResponse struct {
	AccessToken string `json:"accessToken"`
	GetProfileResponse
}

type SetupPasswordResponse struct {
	RetryAfter int64 `json:"retryAfter"`
}
