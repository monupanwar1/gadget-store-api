package dto

import "strings"

type CreateCategoryRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type CategoryResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type CategoryListResponse []CategoryResponse

type UpdateCategoryRequest CreateCategoryRequest

func (req CreateCategoryRequest) Validate() error {

	if strings.TrimSpace(req.Name) == "" {
		return &ValidationError{
			Field: "name",
			Msg:   "must not be empty",
		}
	}

	if len(req.Name) > 200 {
		return &ValidationError{
			Field: "name",
			Msg:   "must be at most 200 characters",
		}
	}

	if strings.TrimSpace(req.Slug) == "" {
		return &ValidationError{
			Field: "slug",
			Msg:   "must not be empty",
		}
	}

	if len(req.Slug) > 200 {
		return &ValidationError{
			Field: "slug",
			Msg:   "must be at most 200 characters",
		}
	}

	return nil

}

func (req UpdateCategoryRequest) Validate() error {

	if strings.TrimSpace(req.Name) == "" {
		return &ValidationError{
			Field: "name",
			Msg:   "must not be empty",
		}
	}

	if len(req.Name) > 200 {
		return &ValidationError{
			Field: "name",
			Msg:   "must be at most 200 characters",
		}
	}

	if strings.TrimSpace(req.Slug) == "" {
		return &ValidationError{
			Field: "slug",
			Msg:   "must not be empty",
		}
	}

	if len(req.Slug) > 200 {
		return &ValidationError{
			Field: "slug",
			Msg:   "must be at most 200 characters",
		}
	}

	return nil
}