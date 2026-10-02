package service

import (
	"crypto/rsa"
	"errors"
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

func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {


	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {

		if token.Method != jwt.SigningMethodRS256 {
			return nil, errors.New("unexpected signing method")
		}
		return s.publicKey, nil
	},
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil

}
