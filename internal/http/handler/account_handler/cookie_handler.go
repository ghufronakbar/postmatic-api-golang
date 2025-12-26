// internal/http/handler/account_handler/cookie_handler.go
package account_handler

import (
	"net/http"
	"strings"
	"time"

	"postmatic-api/config"
)

func SetAuthCookies(w http.ResponseWriter, r *http.Request, cfg *config.Config, accessToken, refreshToken string) {
	domain := cookieDomainForPostmatic(r)
	secure := isSecureRequest(r)

	now := time.Now()
	accessAge := cfg.JWT_ACCESS_TOKEN_EXPIRED
	refreshAge := cfg.JWT_REFRESH_TOKEN_EXPIRED

	accessCookie := &http.Cookie{
		Name:     "postmaticAccessToken",
		Value:    accessToken,
		Path:     "/",
		Domain:   domain,
		HttpOnly: false, // sesuai requirement kamu
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(accessAge.Seconds()),
		Expires:  now.Add(accessAge),
	}

	refreshCookie := &http.Cookie{
		Name:     "postmaticRefreshToken",
		Value:    refreshToken,
		Path:     "/",
		Domain:   domain,
		HttpOnly: true, // ✅ requirement
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(refreshAge.Seconds()),
		Expires:  now.Add(refreshAge),
	}

	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)
}

func isSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		return true
	}
	return false
}

func cookieDomainForPostmatic(r *http.Request) string {
	host := r.Host
	if i := strings.Index(host, ":"); i != -1 {
		host = host[:i]
	}
	h := strings.ToLower(host)

	if h == "localhost" || strings.HasPrefix(h, "localhost") || h == "127.0.0.1" {
		return ""
	}

	if strings.HasSuffix(h, ".postmatic.id") || h == "postmatic.id" {
		return ".postmatic.id"
	}

	return ""
}
