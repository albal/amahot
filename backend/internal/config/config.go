package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DBConnString      string
	Port              string
	ScrapeInterval    time.Duration
	ScrapeURL         string
	BrowserUserAgent  string
}

func Load() Config {
	intervalMins := 30
	if v := os.Getenv("SCRAPE_INTERVAL_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			intervalMins = n
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://amahot:changeme@localhost:5432/amahot?sslmode=disable"
	}

	ua := os.Getenv("SCRAPER_USER_AGENT")
	if ua == "" {
		ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
	}

	return Config{
		DBConnString:     dbURL,
		Port:             port,
		ScrapeInterval:   time.Duration(intervalMins) * time.Minute,
		ScrapeURL:        "https://www.hotukdeals.com/deals?merchant=amazon.co.uk&temperature_filter=100&sortby=hotness",
		BrowserUserAgent: ua,
	}
}
