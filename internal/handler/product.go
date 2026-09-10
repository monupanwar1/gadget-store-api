package handler

import (
	"encoding/json"
	"errors"
	"gadget-store-api/internal/dto"
	"gadget-store-api/internal/httpx"
	"gadget-store-api/internal/model"
	"gadget-store-api/internal/storage"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

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
		httpx.Error(w, http.StatusBadRequest, "invalid form data", httpx.CodeMalformedJSON)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		h.log.Error("invalid product price", "error", err)
		httpx.ValidationError(w, http.StatusBadRequest, "price must be a valid number", httpx.CodeValidationFailed, "price")
		return
	}

	file, header, err := r.FormFile("images")
	if err != nil {
		h.log.Error("failed to get product image", "error", err)
		httpx.ValidationError(w, http.StatusBadRequest, "image is required", httpx.CodeValidationFailed, "image")
		return
	}
	defer file.Close()

	image, err := storage.SaveImage(file, header.Filename)
	if err != nil {
		h.log.Error("failed to save product image", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "failed to save image", httpx.CodeInternalError)
		return
	}

	req := dto.CreateProductRequest{
		Name:        name,
		Description: description,
		Price:       price,
		Image:       image,
	}

	if err := req.Validate(); err != nil {
		var ve *dto.ValidationError
		if errors.As(err, &ve) {
			h.log.Error("validation failed", "field", ve.Field, "msg", ve.Msg)
			httpx.ValidationError(w, http.StatusBadRequest, ve.Msg, httpx.CodeValidationFailed, ve.Field)
			return
		}
		h.log.Error("validation failed", "error", err)
		httpx.Error(w, http.StatusBadRequest, "validation failed", httpx.CodeValidationFailed)
		return
	}

	product := model.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Image:       req.Image,
	}

	if err := h.db.Create(&product).Error; err != nil {
		h.log.Error("failed to create product", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "failed to create product", httpx.CodeInternalError)
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
		h.log.Error("failed to encode product response", "error", err)
	}
}

func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {

	var products []model.Product

	if err := h.db.Find(&products).Error; err != nil {
		h.log.Error("failed to fetch products", "error", err)
		httpx.Error(w, http.StatusInternalServerError, "failed to fetch products", httpx.CodeInternalError)
		return
	}

	h.log.Info("products fetched", "count", len(products))

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
		h.log.Error("failed to encode products response", "error", err)
	}
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	var product model.Product
	if err := h.db.First(&product, id).Error; err != nil {
		h.log.Error("failed to fetch product", "id", id, "error", err)
		httpx.Error(w, http.StatusNotFound, "product not found", httpx.CodeNotFound)
		return
	}

	h.log.Info("product fetched", "id", product.ID)

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
		h.log.Error("failed to encode product response", "error", err)
	}
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	var product model.Product
	if err := h.db.First(&product, id).Error; err != nil {
		h.log.Error("failed to find product for update", "id", id, "error", err)
		httpx.Error(w, http.StatusNotFound, "product not found", httpx.CodeNotFound)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		h.log.Error("failed to parse product form", "error", err)
		httpx.Error(w, http.StatusBadRequest, "invalid form data", httpx.CodeMalformedJSON)
		return
	}

	// Build the DTO from existing values, then overlay only the fields the
	// caller actually sent, so Validate() runs against the full resulting record.
	req := dto.UpdateProductRequest{
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Image:       product.Image,
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
			h.log.Error("invalid product price", "error", err)
			httpx.ValidationError(w, http.StatusBadRequest, "price must be a valid number", httpx.CodeValidationFailed, "price")
			return
		}
		req.Price = value
	}

	if file, fileHeader, err := r.FormFile("images"); err == nil {
		defer file.Close()

		image, err := storage.SaveImage(file, fileHeader.Filename)
		if err != nil {
			h.log.Error("failed to save product image", "error", err)
			httpx.Error(w, http.StatusInternalServerError, "failed to save image", httpx.CodeInternalError)
			return
		}
		req.Image = image
	}

	if err := dto.CreateProductRequest(req).Validate(); err != nil {
		var ve *dto.ValidationError
		if errors.As(err, &ve) {
			h.log.Error("validation failed", "field", ve.Field, "msg", ve.Msg)
			httpx.ValidationError(w, http.StatusBadRequest, ve.Msg, httpx.CodeValidationFailed, ve.Field)
			return
		}
		h.log.Error("validation failed", "error", err)
		httpx.Error(w, http.StatusBadRequest, "validation failed", httpx.CodeValidationFailed)
		return
	}

	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Image = req.Image

	if err := h.db.Save(&product).Error; err != nil {
		h.log.Error("failed to update product", "id", id, "error", err)
		httpx.Error(w, http.StatusInternalServerError, "failed to update product", httpx.CodeInternalError)
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
		h.log.Error("failed to encode product response", "id", id, "error", err)
	}
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	var product model.Product
	if err := h.db.First(&product, id).Error; err != nil {
		h.log.Error("failed to find product for deletion", "id", id, "error", err)
		httpx.Error(w, http.StatusNotFound, "product not found", httpx.CodeNotFound)
		return
	}

	if err := h.db.Delete(&product).Error; err != nil {
		h.log.Error("failed to delete product", "id", id, "error", err)
		httpx.Error(w, http.StatusInternalServerError, "failed to delete product", httpx.CodeInternalError)
		return
	}

	h.log.Info("product deleted", "id", product.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "product deleted successfully",
	})
}
	