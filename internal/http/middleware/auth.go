// internal/http/middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/token"
)

// ContextKey adalah tipe data khusus untuk key context agar tidak bentrok dengan library lain
type contextKey string

const (
	UserContextKey contextKey = "userClaims"
)

// AuthMiddleware adalah middleware utama
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 1. Ekstrak Token dari 3 Sumber
		tokenStr := extractToken(r)

		if tokenStr == "" {
			response.Error(w, errs.NewUnauthorized("MISSING_TOKEN"), nil)
			return
		}

		// 2. Validasi Token menggunakan pkg/token Anda
		claims, err := token.ValidateAccessToken(tokenStr)
		if err != nil {
			// Error bisa karena expired, signature salah, dll
			response.Error(w, errs.NewUnauthorized("INVALID_OR_EXPIRED_TOKEN"), nil)
			return
		}

		// 3. Simpan Claims ke Context
		// Agar bisa diakses di handler selanjutnya (Service/Controller)
		ctx := context.WithValue(r.Context(), UserContextKey, claims)

		// 4. Lanjut ke Next Handler dengan Context baru
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper untuk mengekstrak token berdasarkan prioritas
func extractToken(r *http.Request) string {
	// Prioritas 1: Query Param (?postmaticAccessToken=xxx)
	queryToken := r.URL.Query().Get("postmaticAccessToken")
	if queryToken != "" {
		return queryToken
	}

	// Prioritas 2: Custom Header (X-Postmatic-AccessToken)
	headerToken := r.Header.Get("X-Postmatic-AccessToken")
	if headerToken != "" {
		return headerToken
	}

	// Prioritas 3: Authorization: Bearer xxx
	bearerToken := r.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}

	return ""
}

// --- HELPER UNTUK HANDLER ---

// GetUserFromContext memudahkan pengambilan data user di Controller/Service
func GetUserFromContext(ctx context.Context) (*token.Claims, error) {
	claims, ok := ctx.Value(UserContextKey).(*token.Claims)
	if !ok {
		return nil, errs.NewInternalServerError(nil) // Seharusnya tidak terjadi jika lewat middleware
	}
	return claims, nil
}
