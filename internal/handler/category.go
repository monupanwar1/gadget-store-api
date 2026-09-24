package handler

import (
	"encoding/json"
	"errors"
	"gadget-store-api/internal/dto"
	"gadget-store-api/internal/httpx"
	"gadget-store-api/internal/middleware"
	"gadget-store-api/internal/model"
	"log/slog"
	"net/http"

	"gorm.io/gorm"
)

type CategoryHandler struct {
	log *slog.Logger
	db  *gorm.DB
}

func NewCategoryHandler(log *slog.Logger, db *gorm.DB) *CategoryHandler {
	return &CategoryHandler{
		log: log,
		db:  db,
	}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	var req dto.CreateCategoryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error(
			"failed to decode category request",
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
				"category validation failed",
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
			"category validation failed",
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

	category := model.Category{
		Name: req.Name,
		Slug: req.Slug,
	}

	if err := h.db.WithContext(ctx).Create(&category).Error; err != nil {
		h.log.Error(
			"failed to create category",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
			"failed to create category",
			httpx.CodeInternalError,
		)
		return
	}

	h.log.Info(
		"category created",
		"request_id", requestID,
		"id", category.ID,
	)

	response := dto.CategoryResponse{
		ID:   category.ID,
		Name: category.Name,
		Slug: category.Slug,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error(
			"failed to encode category response",
			"request_id", requestID,
			"error", err,
		)
	}
}

func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	id := r.PathValue("id")

	var category model.Category

	if err := h.db.WithContext(ctx).First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.log.Error(
				"category not found",
				"id", id,
				"request_id", requestID,
			)

			httpx.Error(
				w,
				http.StatusNotFound,
				"category not found",
				httpx.CodeNotFound,
			)
			return
		}

		h.log.Error(
			"failed to fetch category",
			"id", id,
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
			"failed to fetch category",
			httpx.CodeInternalError,
		)
		return
	}

	response := dto.CategoryResponse{
		ID:   category.ID,
		Name: category.Name,
		Slug: category.Slug,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error(
			"failed to encode category response",
			"request_id", requestID,
			"error", err,
		)
	}
}

func (h *CategoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	var categories []model.Category

	if err := h.db.Find(&categories).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.log.Error(
				"failed to  fetch categories",
				"request_id", requestID,
				"error", err,
			)

			httpx.Error(
				w,
				http.StatusNotFound,
				"failed to fetch categories",
				httpx.CodeInternalError,
			)
			return
		}

		response := make(dto.CategoryListResponse, 0, len(categories))

		for _, category := range categories {
			response = append(response, dto.CategoryResponse{
				ID:   category.ID,
				Name: category.Name,
				Slug: category.Slug,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			h.log.Error(
				"failed to encode categories response",
				"request_id", requestID,
				"error", err,
			)
		}
	}
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	id := r.PathValue("id")

	// 1. Check category exists
	var category model.Category

	if err := h.db.WithContext(ctx).First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.log.Error(
				"category not found",
				"id", id,
				"request_id", requestID,
			)

			httpx.Error(
				w,
				http.StatusNotFound,
				"category not found",
				httpx.CodeNotFound,
			)
			return
		}

		h.log.Error(
			"failed to fetch category",
			"id", id,
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
			"failed to fetch category",
			httpx.CodeInternalError,
		)
		return
	}

	// 2. Decode request body
	var req dto.UpdateCategoryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error(
			"failed to decode category request",
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

	// 3. Validate request
	if err := req.Validate(); err != nil {
		var ve *dto.ValidationError

		if errors.As(err, &ve) {
			h.log.Error(
				"category validation failed",
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
			"category validation failed",
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

	// 4. Update category
	category.Name = req.Name
	category.Slug = req.Slug

	// 5. Save changes

	if err := h.db.WithContext(ctx).Save(&category).Error; err != nil {
		h.log.Error(
			"failed to update category",
			"id", category.ID,
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
			"failed to update category",
			httpx.CodeInternalError,
		)
		return
	}

	// 6. Create response
	response := dto.CategoryResponse{
		ID:   category.ID,
		Name: category.Name,
		Slug: category.Slug,
	}

	// 7. Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error(
			"failed to encode category response",
			"id", category.ID,
			"request_id", requestID,
			"error", err,
		)
	}
}
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	id := r.PathValue("id")

	var category model.Category

	// 1. Check category exists
	if err := h.db.WithContext(ctx).First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.log.Error(
				"category not found",
				"id", id,
				"request_id", requestID,
			)

			httpx.Error(
				w,
				http.StatusNotFound,
				"category not found",
				httpx.CodeNotFound,
			)
			return
		}

		h.log.Error(
			"failed to fetch category",
			"id", id,
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
			"failed to fetch category",
			httpx.CodeInternalError,
		)
		return
	}

	// 2. Delete category
	if err := h.db.WithContext(ctx).Delete(&category).Error; err != nil {
		h.log.Error(
			"failed to delete category",
			"id", category.ID,
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
			"failed to delete category",
			httpx.CodeInternalError,
		)
		return
	}

	// 3. Success
	h.log.Info(
		"category deleted",
		"id", category.ID,
		"request_id", requestID,
	)

	w.WriteHeader(http.StatusNoContent)
}
