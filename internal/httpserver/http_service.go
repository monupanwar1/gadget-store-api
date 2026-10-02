package httpserver

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"gadget-store-api/internal/handler"
	"gadget-store-api/internal/middleware"
	"gadget-store-api/internal/service"
)

type HTTPService struct {
	router *chi.Mux
}

func NewHTTPService(
	healthHandler *handler.HealthHandler,
	authHandler *handler.AuthHandler,
	categoryHandler *handler.CategoryHandler,
	productHandler *handler.ProductHandler,
	jwtService *service.JWTService,
) *HTTPService {

	r := chi.NewRouter()

	// =========================
	// Public routes
	// =========================

	r.Get("/healthz", healthHandler.Health)

	r.Post("/login", authHandler.Login)
	r.Post("/register", authHandler.Register)

	r.Get("/categories", categoryHandler.GetAll)
	r.Get("/categories/{id}", categoryHandler.GetByID)

	r.Get("/products", productHandler.GetAll)
	r.Get("/products/{id}", productHandler.GetByID)

	// =========================
	// Authenticated routes
	// =========================

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(jwtService))

		r.Get("/me", authHandler.Me)

		r.Post("/categories", categoryHandler.Create)
		r.Patch("/categories/{id}", categoryHandler.Update)
		r.Delete("/categories/{id}", categoryHandler.Delete)

		r.Post("/products", productHandler.Create)
		r.Patch("/products/{id}", productHandler.Update)
		r.Delete("/products/{id}", productHandler.Delete)
	})

	return &HTTPService{
		router: r,
	}
}

func (s *HTTPService) Handler() http.Handler {
	return s.router
}
