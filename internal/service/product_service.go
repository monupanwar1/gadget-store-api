package service

import (
	"context"
	"errors"
	"gadget-store-api/internal/dto"
	"gadget-store-api/internal/model"

	"gorm.io/gorm"
)

type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{
		db: db,
	}
}

func (s *ProductService) Create(
	ctx context.Context,
	req dto.CreateProductRequest,
) (*model.Product, error) {

	// Check category exists
	var category model.Category

	if err := s.db.WithContext(ctx).
		First(&category, req.CategoryID).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}

		return nil, err
	}

	product := model.Product{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Image:       req.Image,
	}

	if err := s.db.WithContext(ctx).Create(&product).Error; err != nil {
		return nil, err
	}

	return &product, nil
}

// GetAll
func (s *ProductService) GetAll(
	ctx context.Context,
) ([]model.Product, error) {

	var products []model.Product

	if err := s.db.WithContext(ctx).
		Preload("Category").
		Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

// GetByID
func (s *ProductService) GetByID(
	ctx context.Context,
	id string,
) (*model.Product, error) {

	var product model.Product

	if err := s.db.WithContext(ctx).
		Preload("Category").
		First(&product, id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}

		return nil, err
	}

	return &product, nil
}

// GetByCategory
func (s *ProductService) GetByCategory(
	ctx context.Context,
	categoryID string,
) ([]model.Product, error) {
	var products []model.Product

	if err := s.db.WithContext(ctx).
		Preload("Category").
		Where("category_id = ?", categoryID).
		Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

func (s *ProductService) GetAllByCategory(ctx context.Context) ([]model.Product, error) {
	var products []model.Product

	if err := s.db.WithContext(ctx).
		Preload("Category").
		Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

// Update
func (s *ProductService) Update(
	ctx context.Context,
	id string,
	req dto.UpdateProductRequest,
) (*model.Product, error) {

	var product model.Product

	if err := s.db.WithContext(ctx).
		First(&product, id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}

		return nil, err
	}

	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Image = req.Image

	if err := s.db.WithContext(ctx).Save(&product).Error; err != nil {
		return nil, err
	}

	return &product, nil
}

// Delete
func (s *ProductService) Delete(
	ctx context.Context,
	id string,
) error {

	var product model.Product

	if err := s.db.WithContext(ctx).
		First(&product, id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		}

		return err
	}

	if err := s.db.WithContext(ctx).
		Delete(&product).Error; err != nil {
		return err
	}

	return nil
}
