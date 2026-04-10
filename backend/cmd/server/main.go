package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/albal/amahot/backend/internal/api"
	"github.com/albal/amahot/backend/internal/config"
	"github.com/albal/amahot/backend/internal/db"
	"github.com/albal/amahot/backend/internal/repository"
	"github.com/albal/amahot/backend/internal/scraper"
)

func main() {
	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Database
	pool, err := db.NewPool(ctx, cfg.DBConnString)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	// Migrations
	if err := db.RunMigrations(ctx, pool); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	// Repositories
	dealRepo := repository.NewDealRepo(pool)
	clickRepo := repository.NewClickRepo(pool)

	// Scraper
	s := scraper.New(cfg.ScrapeURL, cfg.BrowserUserAgent, dealRepo)
	scraper.Start(ctx, s, cfg.ScrapeInterval)

	// HTTP server
	router := api.NewRouter(dealRepo, clickRepo, pool)
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}
