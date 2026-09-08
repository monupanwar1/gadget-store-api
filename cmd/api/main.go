package main

import (
	"gadget-store-api/internal/config"
	"gadget-store-api/internal/database"
	"gadget-store-api/internal/handler"
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

	if err := os.MkdirAll("logs", 0755); err != nil {
		slog.Error("failed to create logs directory", "error", err)
		os.Exit(1)
	}

	log, logFile, err := logger.New("info", "logs/info.log")

	if err != nil {
		slog.Error("failed to initialize logger", "error", err)
		os.Exit(1)
	}

	defer logFile.Close()

	errorLog, errorFile, err := logger.New("info", "logs/error.log")

	if err != nil {
		slog.Error("failed to initialize  error logger", "error", err)
		os.Exit(1)
	}

	defer errorFile.Close()

	if err := os.MkdirAll("data", 0755); err != nil {
		errorLog.Error("failed to create data directory", "error", err)
		os.Exit(1)
	}

	db, err := database.NewSQLite("data/gadget_store.db")
	if err != nil {
		errorLog.Error("failed to connect database", "error", err)
		os.Exit(1)
	}

	if err := db.AutoMigrate(&model.Product{}); err != nil {
		errorLog.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		errorLog.Error("failed to get database connection", "error", err)
		os.Exit(1)
	}

	defer sqlDB.Close()

	// handler
	healthHandler := handler.NewHealthHandler(log)
	productHandler := handler.NewProductHandler(log, db)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthHandler.Health)
	mux.HandleFunc("POST /products", productHandler.Create)
	mux.HandleFunc("GET /products", productHandler.GetAll)
	mux.HandleFunc("GET /products/{id}", productHandler.GetByID)
	mux.HandleFunc("PATCH /products/{id}", productHandler.Update)
	mux.HandleFunc("DELETE /products/{id}", productHandler.Delete)

	address := ":" + cfg.Port

	server := http.Server{
		Addr:    address,
		Handler: mux,
	}

	log.Info("server started", "port", cfg.Port)

	if err := server.ListenAndServe(); err != nil {
		errorLog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
