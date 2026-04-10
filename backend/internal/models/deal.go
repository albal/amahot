package models

import "time"

type Deal struct {
	ID            int64     `json:"id"`
	ExternalID    string    `json:"external_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Price         string    `json:"price"`
	OriginalPrice string    `json:"original_price"`
	ImageURL      string    `json:"image_url"`
	DealURL       string    `json:"deal_url"`
	Merchant      string    `json:"merchant"`
	Temperature   int       `json:"temperature"`
	Category      string    `json:"category"`
	ScrapedAt     time.Time `json:"scraped_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	IsExpired     bool      `json:"is_expired"`
	ClickCount    int64     `json:"click_count"`
}
