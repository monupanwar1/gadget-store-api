package service

import (
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`

	jwt.RegisteredClaims
}

type JWTService struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewJWTService(
	privateKey *rsa.PrivateKey,
	publicKey *rsa.PublicKey,
) *JWTService {
	return &JWTService{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

func (s *JWTService) GenerateAccessToken(
	userID uint,
	email string,
	role string,
) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "gadget_store_api",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodRS256,
		claims,
	)

	return token.SignedString(s.privateKey)
}
