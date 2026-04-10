package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/albal/amahot/backend/internal/models"
)

type ClickRepo struct {
	pool *pgxpool.Pool
}

func NewClickRepo(pool *pgxpool.Pool) *ClickRepo {
	return &ClickRepo{pool: pool}
}

func (r *ClickRepo) Insert(ctx context.Context, c models.Click) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO clicks (deal_id, ip_address, user_agent, referer)
		VALUES ($1, $2::inet, $3, $4)
	`, c.DealID, c.IPAddress, c.UserAgent, c.Referer)
	if err != nil {
		return fmt.Errorf("insert click for deal %d: %w", c.DealID, err)
	}
	return nil
}
