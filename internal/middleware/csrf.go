package middleware

import (
	"crypto/subtle"
	"net/http"

	"gadget-store-api/internal/httpx"
)

func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Safe methods do not modify server state.
		switch r.Method {
		case http.MethodGet,
			http.MethodHead,
			http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}

		// Get the CSRF token from the cookie.
		cookie, err := r.Cookie(httpx.CSRFTokenCookie)
		if err != nil || cookie.Value == "" {
			httpx.Error(
				w,
				http.StatusForbidden,
				"csrf validation failed",
				httpx.CodeForbidden,
			)
			return
		}

		// Get the CSRF token sent by the frontend.
		headerToken := r.Header.Get("X-CSRF-Token")

		if headerToken == "" ||
			subtle.ConstantTimeCompare(
				[]byte(cookie.Value),
				[]byte(headerToken),
			) != 1 {
			httpx.Error(
				w,
				http.StatusForbidden,
				"csrf validation failed",
				httpx.CodeForbidden,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}
