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
)

type CategoryHandler struct {
	log      *slog.Logger
	category *service.CategoryService
}

func NewCategoryHandler(
	log *slog.Logger,
	category *service.CategoryService,
) *CategoryHandler {
	return &CategoryHandler{
		log:      log,
		category: category,
	}
}

// Create
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

	category, err := h.category.Create(ctx, req)
	if err != nil {
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

// GetByID
func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	id := r.PathValue("id")

	category, err := h.category.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
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

// GetAll
func (h *CategoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	categories, err := h.category.GetAll(ctx)
	if err != nil {
		h.log.Error(
			"failed to fetch categories",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
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

// Update
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	id := r.PathValue("id")

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

	if err := req.Validate(); err != nil {
		var ve *dto.ValidationError

		if errors.As(err, &ve) {
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

	category, err := h.category.Update(ctx, id, req)
	if err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			httpx.Error(
				w,
				http.StatusNotFound,
				"category not found",
				httpx.CodeNotFound,
			)
			return
		}

		h.log.Error(
			"failed to update category",
			"id", id,
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

	h.log.Info(
		"category updated",
		"id", category.ID,
		"request_id", requestID,
	)

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
			"id", category.ID,
			"request_id", requestID,
			"error", err,
		)
	}
}

// Delete
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	id := r.PathValue("id")

	if err := h.category.Delete(ctx, id); err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			httpx.Error(
				w,
				http.StatusNotFound,
				"category not found",
				httpx.CodeNotFound,
			)
			return
		}

		h.log.Error(
			"failed to delete category",
			"id", id,
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

	h.log.Info(
		"category deleted",
		"id", id,
		"request_id", requestID,
	)

	w.WriteHeader(http.StatusNoContent)
}
