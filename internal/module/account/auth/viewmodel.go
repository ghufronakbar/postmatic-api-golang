// internal/module/account/auth/viewmodel.go
package auth

type LoginResponse struct {
	// Profile ID
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	ImageUrl     *string `json:"imageUrl"`
	AccessToken  string  `json:"accessToken"`
	RefreshToken string  `json:"refreshToken"`
}

type RegisterResponse struct {
	// Profile ID
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	ImageUrl   *string `json:"imageUrl"`
	RetryAfter int64   `json:"retryAfter"`
}

type VerifyCreateAccountTokenResponse struct {
	// Profile ID
	ID       *string `json:"id"`
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	ImageUrl *string `json:"imageUrl"`
	Valid    bool    `json:"valid"`
}

type VerifyCreateAccountResponse struct {
	// Profile ID
	ID           *string `json:"id"`
	Name         *string `json:"name"`
	Email        *string `json:"email"`
	ImageUrl     *string `json:"imageUrl"`
	Valid        bool    `json:"valid"`
	AccessToken  string  `json:"accessToken"`
	RefreshToken string  `json:"refreshToken"`
}

type SessionResponse struct {
	// Profile ID
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	ImageUrl     *string `json:"imageUrl"`
	AccessToken  string  `json:"accessToken"`
	RefreshToken string  `json:"refreshToken"`
}

type ResendEmailVerificationResponse struct {
	// Profile ID
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	ImageUrl   *string `json:"imageUrl"`
	RetryAfter int64   `json:"retryAfter"`
}
