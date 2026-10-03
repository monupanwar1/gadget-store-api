package middleware

import (
	"context"
	"net/http"

	"gadget-store-api/internal/httpx"
	"gadget-store-api/internal/service"
)

type contextKey string

const claimsContextKey contextKey = "claims"

func Auth(jwtService *service.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(httpx.AccessTokenCookie)
			if err != nil || cookie.Value == "" {
				httpx.Error(
					w,
					http.StatusUnauthorized,
					"authentication required",
					httpx.CodeUnauthenticated,
				)
				return
			}

			claims, err := jwtService.ValidateAccessToken(cookie.Value)
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

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClaimsFromContext(ctx context.Context) (*service.Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*service.Claims)
	return claims, ok
}
