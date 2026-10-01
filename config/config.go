package config

import (
	"crypto/rsa"
	"encoding/base64"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

func MustLoad() Config {
	if err := godotenv.Load(); err != nil {
		panic("failed to load .env: " + err.Error())
	}

	port := os.Getenv("PORT")
	privateKeyB64 := os.Getenv("JWT_PRIVATE_KEY_B64")
	publicKeyB64 := os.Getenv("JWT_PUBLIC_KEY_B64")

	if port == "" {
		panic("PORT is required")
	}

	if privateKeyB64 == "" {
		panic("JWT_PRIVATE_KEY_B64 is required")
	}

	if publicKeyB64 == "" {
		panic("JWT_PUBLIC_KEY_B64 is required")
	}

	privateKeyPEM, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		panic("failed to decode private key: " + err.Error())
	}

	publicKeyPEM, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		panic("failed to decode public key: " + err.Error())
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic("failed to parse private key: " + err.Error())
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		panic("failed to parse public key: " + err.Error())
	}

	return Config{
		Port:       port,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}
}