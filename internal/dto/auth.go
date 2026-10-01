package dto

import (
	"strings"
	"time"
)

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (req RegisterRequest) Validate() error {

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

	if strings.TrimSpace(req.Email) == "" {
		return &ValidationError{
			Field: "email",
			Msg:   "must not be empty",
		}
	}

	if len(req.Email) > 200 {
		return &ValidationError{
			Field: "email",
			Msg:   "must be at most 200 characters",
		}
	}

	if strings.TrimSpace(req.Password) == "" {
		return &ValidationError{
			Field: "password",
			Msg:   "must be not empty",
		}
	}

	if len(req.Password) > 200 {
		return &ValidationError{
			Field: "Password", Msg: "must be at most 200 characters",
		}
	}

	return nil

}

func (req LoginRequest) Validate() error {

	if strings.TrimSpace(req.Email) == "" {
		return &ValidationError{
			Field: "email",
			Msg:   "must not be empty",
		}
	}

	if len(req.Email) > 200 {
		return &ValidationError{
			Field: "email",
			Msg:   "must be at most 200 characters",
		}
	}

	if strings.TrimSpace(req.Password) == "" {
		return &ValidationError{
			Field: "password",
			Msg:   "must be not empty",
		}
	}

	if len(req.Password) > 200 {
		return &ValidationError{
			Field: "Password", Msg: "must be at most 200 characters",
		}
	}

	return nil

}
