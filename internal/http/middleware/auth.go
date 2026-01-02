// internal/http/middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"postmatic-api/internal/module/headless/token"
	"postmatic-api/internal/repository/entity"
	"postmatic-api/pkg/errs"
	"postmatic-api/pkg/response"
	"postmatic-api/pkg/utils"
)

type contextKey string

const ProfileContextKey contextKey = "profileClaims"

func AuthMiddleware(tm token.TokenMaker, allowedRoles []entity.AppRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractToken(r)
			if tokenStr == "" {
				response.Error(w, r, errs.NewUnauthorized("MISSING_TOKEN"), nil)
				return
			}

			claims, err := tm.ValidateAccessToken(tokenStr)
			if err != nil {
				response.Error(w, r, errs.NewUnauthorized("INVALID_OR_EXPIRED_TOKEN"), nil)
				return
			}

			strAllowedRoles := make([]string, len(allowedRoles))
			for i, role := range allowedRoles {
				strAllowedRoles[i] = string(role)
			}

			if !utils.StringInSlice(string(claims.Role), strAllowedRoles) {
				response.Error(w, r, errs.NewForbidden(""), nil)
				return
			}

			ctx := context.WithValue(r.Context(), ProfileContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	if q := r.URL.Query().Get("postmaticAccessToken"); q != "" {
		return q
	}
	if h := r.Header.Get("X-Postmatic-AccessToken"); h != "" {
		return h
	}
	auth := r.Header.Get("Authorization")
	parts := strings.Split(auth, " ")
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}

func GetProfileFromContext(ctx context.Context) (*token.AccessTokenClaims, error) {
	claims, ok := ctx.Value(ProfileContextKey).(*token.AccessTokenClaims)
	if !ok || claims == nil {
		// lebih cocok 401/forbidden daripada internal error
		return nil, errs.NewUnauthorized("MISSING_AUTH_CONTEXT")
	}
	return claims, nil
}
