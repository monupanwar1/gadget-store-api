package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`

	jwt.RegisteredClaims
}

func GenerateAccessToken(
	userID uint,
	email string,
	role string,
	privateKey *rsa.PrivateKey,
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

	return token.SignedString(privateKey)
}

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return base64.RawURLEncoding.EncodeToString(hash[:])

}
