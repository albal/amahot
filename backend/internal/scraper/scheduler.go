package scraper

import (
	"context"
	"log"
	"time"
)

// Start runs the scraper once after a short startup delay, then on every interval tick.
// It blocks until ctx is cancelled.
func Start(ctx context.Context, s *Scraper, interval time.Duration) {
	go func() {
		// Give the DB/server a moment to fully start before first scrape
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}

		log.Printf("Scheduler: first scrape starting (interval: %s)", interval)
		s.Run(ctx)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Scheduler: shutting down")
				return
			case <-ticker.C:
				log.Println("Scheduler: tick — starting scrape")
				s.Run(ctx)
			}
		}
	}()
}
