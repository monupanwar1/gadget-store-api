package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"gadget-store-api/internal/dto"
	"gadget-store-api/internal/httpx"
	"gadget-store-api/internal/middleware"
	"gadget-store-api/internal/service"
	"gadget-store-api/internal/storage"
)

type ProductHandler struct {
	log     *slog.Logger
	product *service.ProductService
}

func NewProductHandler(
	log *slog.Logger,
	product *service.ProductService,
) *ProductHandler {
	return &ProductHandler{
		log:     log,
		product: product,
	}
}

// Create
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.log.Error(
			"failed to parse product form",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusBadRequest,
			"invalid form data",
			httpx.CodeMalformedJSON,
		)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	categoryID, err := strconv.ParseUint(
		r.FormValue("category_id"),
		10,
		64,
	)
	if err != nil {
		httpx.ValidationError(
			w,
			http.StatusBadRequest,
			"category_id must be a valid number",
			httpx.CodeValidationFailed,
			"category_id",
		)
		return
	}

	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		h.log.Error(
			"invalid product price",
			"request_id", requestID,
			"error", err,
		)

		httpx.ValidationError(
			w,
			http.StatusBadRequest,
			"price must be a valid number",
			httpx.CodeValidationFailed,
			"price",
		)
		return
	}

	file, header, err := r.FormFile("images")
	if err != nil {
		h.log.Error(
			"failed to get product image",
			"request_id", requestID,
			"error", err,
		)

		httpx.ValidationError(
			w,
			http.StatusBadRequest,
			"image is required",
			httpx.CodeValidationFailed,
			"image",
		)
		return
	}
	defer file.Close()

	image, err := storage.SaveImage(file, header.Filename)
	if err != nil {
		h.log.Error(
			"failed to save product image",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
			"failed to save image",
			httpx.CodeInternalError,
		)
		return
	}

	req := dto.CreateProductRequest{
		CategoryID:  uint(categoryID),
		Name:        name,
		Description: description,
		Price:       price,
		Image:       image,
	}

	if err := req.Validate(); err != nil {
		var validationErr *dto.ValidationError

		if errors.As(err, &validationErr) {
			h.log.Error(
				"validation failed",
				"request_id", requestID,
				"field", validationErr.Field,
				"msg", validationErr.Msg,
			)

			httpx.ValidationError(
				w,
				http.StatusBadRequest,
				validationErr.Msg,
				httpx.CodeValidationFailed,
				validationErr.Field,
			)
			return
		}

		h.log.Error(
			"validation failed",
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

	product, err := h.product.Create(ctx, req)
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
			"failed to create product",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
			"failed to create product",
			httpx.CodeInternalError,
		)
		return
	}

	h.log.Info(
		"product created",
		"id", product.ID,
		"request_id", requestID,
	)

	response := dto.ProductResponse{
		ID:           product.ID,
		CategoryID:   product.CategoryID,
		CategoryName: product.Category.Name,
		CategorySlug: product.Category.Slug,
		Name:         product.Name,
		Description:  product.Description,
		Price:        product.Price,
		Image:        product.Image,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error(
			"failed to encode product response",
			"request_id", requestID,
			"error", err,
		)
	}
}

// GetAll
func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	products, err := h.product.GetAll(ctx)
	if err != nil {
		h.log.Error(
			"failed to fetch products",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
			"failed to fetch products",
			httpx.CodeInternalError,
		)
		return
	}

	h.log.Info(
		"products fetched",
		"count", len(products),
		"request_id", requestID,
	)

	response := make(dto.ProductListResponse, 0, len(products))

	for _, product := range products {
		response = append(response, dto.ProductResponse{
			ID:           product.ID,
			CategoryID:   product.CategoryID,
			CategoryName: product.Category.Name,
			CategorySlug: product.Category.Slug,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			Image:        product.Image,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error(
			"failed to encode products response",
			"request_id", requestID,
			"error", err,
		)
	}
}

// GetByID
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	id := r.PathValue("id")

	product, err := h.product.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			httpx.Error(
				w,
				http.StatusNotFound,
				"product not found",
				httpx.CodeNotFound,
			)
			return
		}

		h.log.Error(
			"failed to fetch product",
			"id", id,
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
			"failed to fetch product",
			httpx.CodeInternalError,
		)
		return
	}

	h.log.Info(
		"product fetched",
		"id", product.ID,
		"request_id", requestID,
	)

	response := dto.ProductResponse{
		ID:           product.ID,
		CategoryID:   product.CategoryID,
		CategoryName: product.Category.Name,
		CategorySlug: product.Category.Slug,
		Name:         product.Name,
		Description:  product.Description,
		Price:        product.Price,
		Image:        product.Image,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error(
			"failed to encode product response",
			"request_id", requestID,
			"error", err,
		)
	}
}

// GetByCategory
func (h *ProductHandler) GetByCategory(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	categoryID := r.PathValue("categoryID")

	products, err := h.product.GetByCategory(ctx, categoryID)
	if err != nil {
		h.log.Error(
			"failed to get products by category",
			"request_id", requestID,
			"category_id", categoryID,
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

	if len(products) == 0 {
		httpx.Error(
			w,
			http.StatusNotFound,
			"no products found for category",
			httpx.CodeNotFound,
		)
		return
	}

	response := dto.ProductsByCategoryResponse{
		CategoryID:   products[0].CategoryID,
		CategoryName: products[0].Category.Name,
		CategorySlug: products[0].Category.Slug,
		Products:     make([]dto.ProductItemResponse, 0, len(products)),
	}

	for _, product := range products {
		response.Products = append(
			response.Products,
			dto.ProductItemResponse{
				ID:          product.ID,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
				Image:       product.Image,
			},
		)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error(
			"failed to encode products by category response",
			"request_id", requestID,
			"error", err,
		)
	}
}


func (h *ProductHandler) GetAllByCategory(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	products, err := h.product.GetAllByCategory(ctx)
	if err != nil {
		h.log.Error(
			"failed to get products by category",
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

	categoryMap := make(map[uint]*dto.ProductsByCategoryResponse)

	for _, product := range products {
		category, exists := categoryMap[product.CategoryID]

		if !exists {
			category = &dto.ProductsByCategoryResponse{
				CategoryID:   product.CategoryID,
				CategoryName: product.Category.Name,
				CategorySlug: product.Category.Slug,
				Products:     make([]dto.ProductItemResponse, 0),
			}

			categoryMap[product.CategoryID] = category
		}

		category.Products = append(
			category.Products,
			dto.ProductItemResponse{
				ID:          product.ID,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
				Image:       product.Image,
			},
		)
	}

	response := make([]dto.ProductsByCategoryResponse, 0, len(categoryMap))

	for _, category := range categoryMap {
		response = append(response, *category)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error(
			"failed to encode products by category response",
			"request_id", requestID,
			"error", err,
		)
	}
}

// Update
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	id := r.PathValue("id")

	// Get existing product through service
	product, err := h.product.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			httpx.Error(
				w,
				http.StatusNotFound,
				"product not found",
				httpx.CodeNotFound,
			)
			return
		}

		h.log.Error(
			"failed to find product for update",
			"id", id,
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
			"failed to fetch product",
			httpx.CodeInternalError,
		)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.log.Error(
			"failed to parse product form",
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusBadRequest,
			"invalid form data",
			httpx.CodeMalformedJSON,
		)
		return
	}

	// Start with existing values
	req := dto.UpdateProductRequest{
		CategoryID:  product.CategoryID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Image:       product.Image,
	}

	// Overlay fields sent by client
	if categoryID := strings.TrimSpace(r.FormValue("category_id")); categoryID != "" {
		value, err := strconv.ParseUint(categoryID, 10, 64)
		if err != nil {
			httpx.ValidationError(
				w,
				http.StatusBadRequest,
				"category_id must be a valid number",
				httpx.CodeValidationFailed,
				"category_id",
			)
			return
		}

		req.CategoryID = uint(value)
	}

	if name := r.FormValue("name"); strings.TrimSpace(name) != "" {
		req.Name = name
	}

	if description := r.FormValue("description"); strings.TrimSpace(description) != "" {
		req.Description = description
	}

	if priceStr := r.FormValue("price"); priceStr != "" {
		value, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			h.log.Error(
				"invalid product price",
				"request_id", requestID,
				"error", err,
			)

			httpx.ValidationError(
				w,
				http.StatusBadRequest,
				"price must be a valid number",
				httpx.CodeValidationFailed,
				"price",
			)
			return
		}

		req.Price = value
	}

	if file, fileHeader, err := r.FormFile("images"); err == nil {
		defer file.Close()

		image, err := storage.SaveImage(file, fileHeader.Filename)
		if err != nil {
			h.log.Error(
				"failed to save product image",
				"request_id", requestID,
				"error", err,
			)

			httpx.Error(
				w,
				http.StatusInternalServerError,
				"failed to save image",
				httpx.CodeInternalError,
			)
			return
		}

		req.Image = image
	}

	if err := dto.CreateProductRequest(req).Validate(); err != nil {
		var validationErr *dto.ValidationError

		if errors.As(err, &validationErr) {
			httpx.ValidationError(
				w,
				http.StatusBadRequest,
				validationErr.Msg,
				httpx.CodeValidationFailed,
				validationErr.Field,
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

	product, err = h.product.Update(ctx, id, req)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			httpx.Error(
				w,
				http.StatusNotFound,
				"product not found",
				httpx.CodeNotFound,
			)
			return
		}

		h.log.Error(
			"failed to update product",
			"id", id,
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
			"failed to update product",
			httpx.CodeInternalError,
		)
		return
	}

	h.log.Info(
		"product updated",
		"id", product.ID,
		"request_id", requestID,
	)

	response := dto.ProductResponse{
		ID:          product.ID,
		CategoryID:  product.CategoryID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Image:       product.Image,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error(
			"failed to encode product response",
			"id", id,
			"request_id", requestID,
			"error", err,
		)
	}
}

// Delete
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.RequestIDFromContext(ctx)

	id := r.PathValue("id")

	if err := h.product.Delete(ctx, id); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			httpx.Error(
				w,
				http.StatusNotFound,
				"product not found",
				httpx.CodeNotFound,
			)
			return
		}

		h.log.Error(
			"failed to delete product",
			"id", id,
			"request_id", requestID,
			"error", err,
		)

		httpx.Error(
			w,
			http.StatusInternalServerError,
			"failed to delete product",
			httpx.CodeInternalError,
		)
		return
	}

	h.log.Info(
		"product deleted",
		"id", id,
		"request_id", requestID,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "product deleted successfully",
	})
}
