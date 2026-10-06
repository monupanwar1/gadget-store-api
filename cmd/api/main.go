package main

import (
	"errors"
	"gadget-store-api/config"
	"gadget-store-api/internal/database"
	"gadget-store-api/internal/handler"
	"gadget-store-api/internal/httpserver"
	"gadget-store-api/internal/logger"
	"gadget-store-api/internal/middleware"
	"gadget-store-api/internal/service"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	cfg := config.MustLoad()

	// Create logs directory
	if err := os.MkdirAll("logs", 0755); err != nil {
		slog.Error("failed to create logs directory", "error", err)
		os.Exit(1)
	}

	// Application logger
	log, logFile, err := logger.New("info", "logs/info.log")
	if err != nil {
		slog.Error("failed to initialize logger", "error", err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Error logger
	errorLog, errorFile, err := logger.New("info", "logs/error.log")
	if err != nil {
		slog.Error("failed to initialize error logger", "error", err)
		os.Exit(1)
	}
	defer errorFile.Close()

	// Create data directory
	if err := os.MkdirAll("data", 0755); err != nil {
		errorLog.Error("failed to create data directory", "error", err)
		os.Exit(1)
	}

	// Database
	db, err := database.NewSQLite("data/gadget_store.db")
	if err != nil {
		errorLog.Error("failed to connect database", "error", err)
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		errorLog.Error("failed to get database connection", "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	// JWT service
	jwtService := service.NewJWTService(
		cfg.PrivateKey,
		cfg.PublicKey,
	)

	// Auth service
	authService := service.NewAuthService(
		db,
		jwtService,
	)

	categoryService := service.NewCategoryService(
		db,
	)

	productService := service.NewProductService(
		db,
	)

	// Handlers
	authHandler := handler.NewAuthHandler(
		log,
		authService,
	)

	healthHandler := handler.NewHealthHandler(log)

	categoryHandler := handler.NewCategoryHandler(
		log,
		categoryService,
	)

	productHandler := handler.NewProductHandler(
		log,
		productService,
	)

	// HTTP service / Chi router
	httpService := httpserver.NewHTTPService(
		healthHandler,
		authHandler,
		categoryHandler,
		productHandler,
		jwtService,
	)

	// Global middleware
	httpHandler := middleware.RequestId(
		httpService.Handler(),
	)

	// HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: httpHandler,

		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Info("server started", "port", cfg.Port)

	if err := server.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		errorLog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
