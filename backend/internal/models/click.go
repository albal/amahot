package models

import "time"

type Click struct {
	ID        int64     `json:"id"`
	DealID    int64     `json:"deal_id"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Referer   string    `json:"referer"`
	ClickedAt time.Time `json:"clicked_at"`
}
