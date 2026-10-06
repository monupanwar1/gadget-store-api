package dto

import (
	"fmt"
	"strings"
)

type CreateProductRequest struct {
	CategoryID  uint    `json:"category_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
}

type ProductResponse struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	Image        string  `json:"image"`
	CategoryID   uint    `json:"category_id"`
	CategoryName string  `json:"category_name"`
	CategorySlug string  `json:"category_slug"`
}

type ProductItemResponse struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
}

type ProductsByCategoryResponse struct {
	CategoryID   uint                  `json:"category_id"`
	CategoryName string                `json:"category_name"`
	CategorySlug string                `json:"category_slug"`
	Products     []ProductItemResponse `json:"products"`
}

type ProductListResponse []ProductResponse
type UpdateProductRequest CreateProductRequest

type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s:%s", e.Field, e.Msg)
}

func (req CreateProductRequest) Validate() error {

	if req.CategoryID == 0 {
		return &ValidationError{
			Field: "category_id",
			Msg:   "must be greater than zero",
		}
	}

	if strings.TrimSpace(req.Name) == "" {
		return &ValidationError{
			Field: "title",
			Msg:   "Must not be empty",
		}
	}
	if len(req.Name) > 200 {
		return &ValidationError{Field: "title", Msg: "must be at most 200 characters"}
	}
	if strings.TrimSpace(req.Description) == "" {
		return &ValidationError{Field: "description", Msg: "must not be empty"}
	}
	if len(req.Description) > 5000 {
		return &ValidationError{Field: "description", Msg: "must be at most 5000 characters"}
	}
	if req.Price <= 0 {
		return &ValidationError{Field: "price", Msg: "must be greater than zero"}
	}
	if strings.TrimSpace(req.Image) == "" {
		return &ValidationError{Field: "city", Msg: "Image not be empty"}
	}
	return nil
}
