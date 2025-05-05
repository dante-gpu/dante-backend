package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dante-gpu/dante-backend/api-gateway/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// I should load the configuration first.
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// I need to set up the router.
	r := chi.NewRouter()

	// I should add some basic middleware.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger) // I'll use Zap logger later for structured logging.
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.RequestTimeout))

	// I'll define a simple health check endpoint for now.
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "API Gateway is healthy")
	})

	// I need to start the HTTP server.
	log.Printf("Starting API Gateway on port %s", cfg.Port)
	if err := http.ListenAndServe(cfg.Port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
