package main

import (
	"encoding/json"
	"gadget-store-api/internal/config"
	"gadget-store-api/internal/database"
	"gadget-store-api/internal/logger"
	"gadget-store-api/internal/model"
	"log/slog"
	"net/http"
	"os"
)

type HealthResponse struct {
	Status string `json:"status"`
}

func main() {
	cfg := config.Load()
	log, logFile, err := logger.New("info", "")

	if err != nil {
		slog.Error("failed to initialize logger", "error", err)
		os.Exit(1)
	}

	if logFile != nil {
		defer logFile.Close()
	}

	if err := os.MkdirAll("data", 0755); err != nil {
		log.Error("failed to create data directory", "error", err)
		os.Exit(1)
	}

	db, err := database.NewSQLite("data/gadget_store.db")
	if err != nil {
		log.Error("failed to connect database", "error", err)
		os.Exit(1)
	}

	if err := db.AutoMigrate(&model.Product{}); err != nil {
		log.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Error("failed to get database connection", "error", err)
		os.Exit(1)
	}

	defer sqlDB.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		response := HealthResponse{
			Status: "ok",
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Error("failed to encode health response", "error", err)
		}
	})

	address := ":" + cfg.Port

	server := http.Server{
		Addr:    address,
		Handler: mux,
	}

	log.Info("Server started", "port", cfg.Port)

	if err := server.ListenAndServe(); err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
