package handler

import (
	"encoding/json"
	"gadget-store-api/internal/dto"
	"gadget-store-api/internal/httpx"
	"gadget-store-api/internal/model"
	"gadget-store-api/internal/storage"
	"log/slog"
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

type ProductHandler struct {
	log *slog.Logger
	db  *gorm.DB
}

func NewProductHandler(log *slog.Logger, db *gorm.DB) *ProductHandler {
	return &ProductHandler{
		log: log,
		db:  db,
	}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.log.Error("failed to parse product form", "error", err)

		httpx.ErrorResponse(
			w,
			http.StatusBadRequest,
			"invalid form data",
			"INVALID_REQUEST",
		)
		return

	}

	name := r.FormValue("name")
	description := r.FormValue("description")
	price, err := strconv.ParseFloat(r.FormValue("price"), 64)

	if err != nil {
		h.log.Error("invalid product price", "error", err)

		httpx.ErrorResponse(
			w,
			http.StatusBadRequest,
			"invalid price",
			"INVALID_PRICE",
		)
		return
	}

	file, header, err := r.FormFile("images")
	if err != nil {
		h.log.Error("failed to get product image", "error", err)

		httpx.ErrorResponse(
			w,
			http.StatusBadRequest,
			"image is required",
			"IMAGE_REQUIRED",
		)
		return
	}
	defer file.Close()

	image, err := storage.SaveImage(file, header.Filename)
	if err != nil {
		h.log.Error("failed to save product image", "error", err)

		httpx.ErrorResponse(
			w,
			http.StatusBadRequest,
			"failed to save image",
			"INTERNAL_ERROR",
		)
		return
	}

	product := model.Product{
		Name:        name,
		Description: description,
		Price:       price,
		Image:       image,
	}

	if err := h.db.Create(&product).Error; err != nil {
		h.log.Error("failed to create product", "error", err)
		httpx.ErrorResponse(w, http.StatusInternalServerError, "failed to create product", "INTERNAL_ERROR")
		return
	}
	h.log.Info("product created", "id", product.ID)

	response := dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Image:       product.Image,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("Failed to encode product response", "error", err)
	}

}

func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {

	var products []model.Product

	if err := h.db.Find(&products).Error; err != nil {
		h.log.Error("failed to fetch products", "error", err)

		httpx.ErrorResponse(w, http.StatusInternalServerError, "failed to fetch products", "INTERNAL_ERROR")
		return
	}

	h.log.Info("Products fetched", "count", len(products))

	response := make(dto.ProductListResponse, 0, len(products))

	for _, product := range products {
		response = append(response, dto.ProductResponse{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Image:       product.Image,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("Failed to encode products response", "error", err)
	}

}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	var product model.Product

	if err := h.db.First(&product, id).Error; err != nil {
		h.log.Error("failed to fetch product", "id", id, "error", err)

		httpx.ErrorResponse(
			w,
			http.StatusNotFound,
			"product not found",
			"PRODUCT_NOT_FOUND",
		)
		return
	}

	h.log.Info("Product fetched", "id", product.ID)

	response := dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Image:       product.Image,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("Failed to encode product response", "error", err)
	}

}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	var product model.Product

	if err := h.db.First(&product, id).Error; err != nil {
		h.log.Error("failed to find product for update", "id", id, "error", err)

		httpx.ErrorResponse(w, http.StatusNotFound, "product not found",
			"PRODUCT_NOT_FOUND")

		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.log.Error("Failed to parse product form", "error", err)

		httpx.ErrorResponse(w, http.StatusBadRequest, "invalid form data", "INVALID_REQUEST")
		return

	}

	name := r.FormValue("name")
	if name != "" {
		product.Name = name
	}

	description := r.FormValue("description")
	if description != "" {
		product.Description = description
	}

	price := r.FormValue("price")

	if price != "" {
		value, err := strconv.ParseFloat(price, 64)
		if err != nil {
			h.log.Error("invalid product price", "error", err)
			httpx.ErrorResponse(w, http.StatusBadRequest, "invalid price", "INVALID_PRICE")
			return
		}
		product.Price = value
	}

	file, handler, err := r.FormFile("images")

	if err == nil {
		defer file.Close()

		image, err := storage.SaveImage(file, handler.Filename)

		if err != nil {
			h.log.Error("failed to save product image", "error", err)
			httpx.ErrorResponse(w, http.StatusInternalServerError, "failed to save image", "INTERNAL_ERROR")
			return
		}
		product.Image = image

	}

	if err := h.db.Save(&product).Error; err != nil {
		h.log.Error("failed to update product", "id", id, "error", err)

		httpx.ErrorResponse(
			w,
			http.StatusInternalServerError,
			"failed to update product",
			"INTERNAL_ERROR",
		)
		return
	}
	h.log.Info("product updated", "id", product.ID)

	response := dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Image:       product.Image,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("Failed to encode product response", "id", id, "error", err)
	}

}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	var product model.Product

	if err := h.db.First(&product, id).Error; err != nil {
		h.log.Error("failed to find product for deletion ", "id", id, "error", err)

		httpx.ErrorResponse(w, http.StatusNotFound, "product not found", "PRODUCT_NOT_FOUND")
		return
	}

	if err := h.db.Delete(&product).Error; err != nil {
		h.log.Error("failed to delete product", "id", id, "error", err)

		httpx.ErrorResponse(w, http.StatusInternalServerError, "failed to delete product", "INTERNAL_ERROR")
		return
	}

	h.log.Info("Product Deleted", "id", product.ID)

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "product deleted successfully",
	})

}
