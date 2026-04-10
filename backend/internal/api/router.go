package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"

	"github.com/albal/amahot/backend/internal/repository"
)

type healthChecker interface {
	Ping(ctx context.Context) error
}

func NewRouter(
	dealRepo *repository.DealRepo,
	clickRepo *repository.ClickRepo,
	db healthChecker,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Restrict CORS to the production origin. Allow localhost for local dev.
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "https://amahot.tsew.com"
	}
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{allowedOrigin, "http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: false,
	})
	r.Use(c.Handler)

	dealsH := NewDealsHandler(dealRepo)
	clickH := NewClickHandler(dealRepo, clickRepo)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "ok"
		if err := db.Ping(r.Context()); err != nil {
			dbStatus = "error"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"db":     dbStatus,
		})
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/deals", dealsH.List)
		r.Get("/deals/{id}", dealsH.GetOne)
		r.Post("/clicks/{id}", clickH.Record)
	})

	return r
}
