package handler

import (
	"encoding/json"
	"gadget-store-api/internal/dto"
	"gadget-store-api/internal/httpx"
	"gadget-store-api/internal/model"
	"log/slog"
	"net/http"

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

	var req dto.CreateProductRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("failed to decode product request", "error", err)

		httpx.ErrorResponse(w, http.StatusBadRequest, "invalid request body", "INVALID_REQUEST")
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
