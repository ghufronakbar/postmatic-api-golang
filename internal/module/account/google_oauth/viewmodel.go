// internal/module/account/google_oauth/viewmodel.go
package google_oauth

type LoginGoogleResponse struct {
	// Profile ID
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	ImageUrl     *string `json:"imageUrl"`
	AccessToken  string  `json:"accessToken"`
	RefreshToken string  `json:"refreshToken"`
	From         string  `json:"from"`
}

// Response untuk endpoint "ambil auth url" (dipakai tombol FE)
type GoogleOAuthAuthURLResponse struct {
	AuthURL string `json:"authUrl"`
}
