package handler

import (
	"encoding/json"
	"errors"
	"gadget-store-api/internal/dto"
	"gadget-store-api/internal/httpx"
	"gadget-store-api/internal/middleware"
	"gadget-store-api/internal/model"
	"gadget-store-api/internal/service"
	"gadget-store-api/internal/utils"
	"log/slog"
	"net/http"

	"gorm.io/gorm"
)

type AuthHandler struct {
	log *slog.Logger
	db  *gorm.DB
	jwt *service.JWTService
}

func NewAuthHandler(log *slog.Logger, db *gorm.DB, jwt *service.JWTService) *AuthHandler {
	return &AuthHandler{
		log: log,
		db:  db,
		jwt: jwt,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	var req dto.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error(
			"failed to decode Regsiter request",
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
				"Register validation failed",
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

		h.log.Error(
			"Regsiter validation failed",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusBadRequest,
			"validation failed",
			httpx.CodeValidationFailed,
		)
		return
	}

	// hash the password

	hashedPassword, err := utils.HashPassword(req.Password)

	if err != nil {
		h.log.Error(
			"failed to hash password for Regsiter request",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
			"invalid request body",
			httpx.CodeMalformedJSON,
		)
		return
	}
	// create user

	user := model.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	err = h.db.WithContext(ctx).Create(&user).Error
	if err != nil {
		h.log.Error(
			"failed to Register",
			"request_id", requestID,
			"error", err,
		)
		httpx.Error(
			w,
			http.StatusInternalServerError,
			"failed to create user",
			httpx.CodeInternalError,
		)
	}
	h.log.Info("User Regsitered ", "id", user.ID)

	if err != nil {
		h.log.Error(
			"failed to Register",
			"request_id", requestID,
			"error", err,
		)
		return
	}

	response := &dto.RegisterResponse{
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(response)
	if err != nil {
		h.log.Error(
			"failed to encode register response",
			"request_id", requestID,
			"error", err,
		)
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
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

	var user model.User

	if err := h.db.WithContext(ctx).
		Select("id", "email", "password_hash", "role").
		Where("email = ?", req.Email).
		First(&user).Error; err != nil {

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
			"failed to find user",
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

	if err := utils.ComparePassword(
		req.Password,
		user.PasswordHash,
	); err != nil {
		httpx.Error(
			w,
			http.StatusUnauthorized,
			"invalid email or password",
			httpx.CodeUnauthenticated,
		)
		return
	}

	accessToken, err := h.jwt.GenerateAccessToken(
		user.ID,
		user.Email,
		user.Role,
	)

	if err != nil {
		h.log.Error(
			"failed to generate access token",
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

	response := dto.LoginResponse{
		AccessToken: accessToken,
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
