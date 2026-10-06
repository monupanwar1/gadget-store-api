package service

import (
	"context"
	"errors"

	"gadget-store-api/internal/dto"
	"gadget-store-api/internal/model"

	"gorm.io/gorm"
)

type CategoryService struct {
	db *gorm.DB
}

func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{
		db: db,
	}
}

// Create
func (s *CategoryService) Create(
	ctx context.Context,
	req dto.CreateCategoryRequest,
) (*model.Category, error) {

	category := model.Category{
		Name: req.Name,
		Slug: req.Slug,
	}

	if err := s.db.WithContext(ctx).
		Create(&category).Error; err != nil {
		return nil, err
	}

	return &category, nil
}

// GetByID
func (s *CategoryService) GetByID(
	ctx context.Context,
	id string,
) (*model.Category, error) {

	var category model.Category

	if err := s.db.WithContext(ctx).
		First(&category, id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}

		return nil, err
	}

	return &category, nil
}

// GetAll
func (s *CategoryService) GetAll(
	ctx context.Context,
) ([]model.Category, error) {

	var categories []model.Category

	if err := s.db.WithContext(ctx).
		Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

// Update
func (s *CategoryService) Update(
	ctx context.Context,
	id string,
	req dto.UpdateCategoryRequest,
) (*model.Category, error) {

	var category model.Category

	if err := s.db.WithContext(ctx).
		First(&category, id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}

		return nil, err
	}

	category.Name = req.Name
	category.Slug = req.Slug

	if err := s.db.WithContext(ctx).
		Save(&category).Error; err != nil {
		return nil, err
	}

	return &category, nil
}

// Delete
func (s *CategoryService) Delete(
	ctx context.Context,
	id string,
) error {

	var category model.Category

	if err := s.db.WithContext(ctx).
		First(&category, id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}

		return err
	}

	if err := s.db.WithContext(ctx).
		Delete(&category).Error; err != nil {
		return err
	}

	return nil
}
