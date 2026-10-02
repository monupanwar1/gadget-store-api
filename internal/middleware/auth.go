package middleware

import (
	"context"
	"net/http"
	"strings"

	"gadget-store-api/internal/httpx"
	"gadget-store-api/internal/service"
)

type contextKey string

const claimsContextKey contextKey = "claims"

func Auth(jwtService *service.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				httpx.Error(
					w,
					http.StatusUnauthorized,
					"authentication required",
					httpx.CodeUnauthenticated,
				)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)

			if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
				httpx.Error(
					w,
					http.StatusUnauthorized,
					"invalid authorization header",
					httpx.CodeUnauthenticated,
				)
				return
			}

			tokenString := parts[1]

			claims, err := jwtService.ValidateAccessToken(tokenString)
			if err != nil {
				httpx.Error(
					w,
					http.StatusUnauthorized,
					"invalid or expired token",
					httpx.CodeUnauthenticated,
				)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				claimsContextKey,
				claims,
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		})
	}
}

func ClaimsFromContext(
	ctx context.Context,
) (*service.Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*service.Claims)

	return claims, ok
}
