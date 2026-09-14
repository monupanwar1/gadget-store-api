package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
}

func MustLoad() Config {

	godotenv.Load()

	port := os.Getenv("PORT")

	if port == "" {
		panic("PORT is required")
	}

	return Config{
		Port: port,
	}
}
