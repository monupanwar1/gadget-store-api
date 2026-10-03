package httpx

import "net/http"

const (
	AccessTokenCookie  = "access_token"
	RefreshTokenCookie = "refresh_token"
	CSRFTokenCookie    = "csrf_token"
)

const (
	accessTokenMaxAge  = 15 * 60
	refreshTokenMaxAge = 7 * 24 * 60 * 60
	csrfTokenMaxAge    = 7 * 24 * 60 * 60
)

func SetAuthCookie(w http.ResponseWriter,
	accessToken string,
	refreshToken string,
	csrfToken string) {

	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   accessTokenMaxAge,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   refreshTokenMaxAge,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     CSRFTokenCookie,
		Value:    csrfToken,
		Path:     "/",
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   csrfTokenMaxAge,
	})
}

func ClearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     CSRFTokenCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

}
