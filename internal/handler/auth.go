package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"gadget-store-api/internal/dto"
	"gadget-store-api/internal/httpx"
	"gadget-store-api/internal/middleware"
	"gadget-store-api/internal/service"
	"gadget-store-api/internal/utils"

	"gorm.io/gorm"
)

type AuthHandler struct {
	log  *slog.Logger
	auth *service.AuthService
}

func NewAuthHandler(
	log *slog.Logger,
	auth *service.AuthService,
) *AuthHandler {
	return &AuthHandler{
		log:  log,
		auth: auth,
	}
}

func (h *AuthHandler) Register(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	var req dto.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error(
			"failed to decode register request",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusBadRequest,
			"invalid request body",
			httpx.CodeMalformedJSON,
		)
		return
	}

	if err := req.Validate(); err != nil {
		var ve *dto.ValidationError

		if errors.As(err, &ve) {
			h.log.Error(
				"register validation failed",
				"request_id", requestID,
				"field", ve.Field,
				"error", ve.Msg,
			)

			httpx.ValidationError(
				w,
				http.StatusBadRequest,
				ve.Msg,
				httpx.CodeValidationFailed,
				ve.Field,
			)
			return
		}

		httpx.Error(
			w,
			http.StatusBadRequest,
			"validation failed",
			httpx.CodeValidationFailed,
		)
		return
	}

	user, err := h.auth.Register(ctx, req)
	if err != nil {
		h.log.Error(
			"failed to register user",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
			"failed to create user",
			httpx.CodeInternalError,
		)
		return
	}

	h.log.Info(
		"user registered",
		"request_id", requestID,
		"user_id", user.ID,
	)

	response := dto.RegisterResponse{
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error(
			"failed to encode register response",
			"request_id", requestID,
			"error", err,
		)
	}
}

func (h *AuthHandler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	var req dto.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error(
			"failed to decode login request",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusBadRequest,
			"invalid request body",
			httpx.CodeMalformedJSON,
		)
		return
	}

	if err := req.Validate(); err != nil {
		var ve *dto.ValidationError

		if errors.As(err, &ve) {
			h.log.Error(
				"login validation failed",
				"request_id", requestID,
				"field", ve.Field,
				"error", ve.Msg,
			)

			httpx.ValidationError(
				w,
				http.StatusBadRequest,
				ve.Msg,
				httpx.CodeValidationFailed,
				ve.Field,
			)
			return
		}

		httpx.Error(
			w,
			http.StatusBadRequest,
			"validation failed",
			httpx.CodeValidationFailed,
		)
		return
	}

	accessToken, refreshToken, err := h.auth.Login(ctx, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpx.Error(
				w,
				http.StatusUnauthorized,
				"invalid email or password",
				httpx.CodeUnauthenticated,
			)
			return
		}

		h.log.Error(
			"failed to login",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
			"internal server error",
			httpx.CodeInternalError,
		)
		return
	}

	csrfToken, err := utils.GenerateCSRFToken()
	if err != nil {
		h.log.Error(
			"failed to generate csrf token",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
			"internal server error",
			httpx.CodeInternalError,
		)
		return
	}

	httpx.SetAuthCookie(
		w,
		accessToken,
		refreshToken,
		csrfToken,
	)

	h.log.Info(
		"user logged in",
		"request_id", requestID,
	)

	response := dto.LoginResponse{
		Message: "login successful",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error(
			"failed to encode login response",
			"request_id", requestID,
			"error", err,
		)
	}
}

func (h *AuthHandler) Me(
	w http.ResponseWriter,
	r *http.Request,
) {
	claims, ok := middleware.ClaimsFromContext(r.Context())

	if !ok {
		httpx.Error(
			w,
			http.StatusUnauthorized,
			"authentication required",
			httpx.CodeUnauthenticated,
		)
		return
	}

	response := dto.MeResponse{
		ID:    claims.UserID,
		Email: claims.Email,
		Role:  claims.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error(
			"failed to encode me response",
			"error", err,
		)
	}
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	cookie, err := r.Cookie(httpx.RefreshTokenCookie)
	if err != nil || cookie.Value == "" {
		h.log.Error(
			"refresh token missing during logout",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusUnauthorized,
			"refresh token required",
			httpx.CodeUnauthenticated,
		)
		return
	}

	if err := h.auth.Logout(ctx, cookie.Value); err != nil {
		h.log.Error(
			"failed to logout user",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
			"internal server error",
			httpx.CodeInternalError,
		)
		return
	}

	httpx.ClearAuthCookie(w)

	h.log.Info(
		"user logged out",
		"request_id", requestID,
	)

	response := dto.LoginResponse{
		Message: "logout successful",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error(
			"failed to encode logout response",
			"request_id", requestID,
			"error", err,
		)
	}
}
