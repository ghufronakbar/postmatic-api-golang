package config

import "golang.org/x/oauth2"

func (c *Config) GoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.OAUTH_GOOGLE_CLIENT_ID,
		ClientSecret: c.OAUTH_GOOGLE_CLIENT_SECRET,
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://oauth2.googleapis.com/token",
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
		},
		RedirectURL: c.API_URL + c.OAUTH_GOOGLE_REDIRECT_URL,
		Scopes: []string{
			"openid",
			"email",
			"profile",
		},
	}
}
