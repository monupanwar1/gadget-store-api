package service

import (
	"context"
	"errors"
	"time"

	"gadget-store-api/internal/dto"
	"gadget-store-api/internal/model"
	"gadget-store-api/internal/utils"

	"gorm.io/gorm"
)

type AuthService struct {
	db  *gorm.DB
	jwt *JWTService
}

func NewAuthService(
	db *gorm.DB,
	jwt *JWTService,
) *AuthService {
	return &AuthService{
		db:  db,
		jwt: jwt,
	}
}

func (s *AuthService) Register(
	ctx context.Context,
	req dto.RegisterRequest,
) (*model.User, error) {

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := model.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) Login(
	ctx context.Context,
	req dto.LoginRequest,
) (string, string, error) {

	var user model.User

	if err := s.db.WithContext(ctx).
		Select("id", "email", "password_hash", "role").
		Where("email = ?", req.Email).
		First(&user).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", gorm.ErrRecordNotFound
		}

		return "", "", err
	}

	if err := utils.ComparePassword(
		req.Password,
		user.PasswordHash,
	); err != nil {
		return "", "", err
	}

	accessToken, err := s.jwt.GenerateAccessToken(
		user.ID,
		user.Email,
		user.Role,
	)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.CreateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) CreateRefreshToken(
	userID uint,
) (string, error) {

	token, err := utils.GenerateRefreshToken()
	if err != nil {
		return "", err
	}

	tokenHash := utils.HashRefreshToken(token)

	refreshToken := model.RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.db.Create(&refreshToken).Error; err != nil {
		return "", err
	}

	return token, nil
}