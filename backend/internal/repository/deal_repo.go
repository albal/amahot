package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/albal/amahot/backend/internal/models"
)

type DealRepo struct {
	pool *pgxpool.Pool
}

func NewDealRepo(pool *pgxpool.Pool) *DealRepo {
	return &DealRepo{pool: pool}
}

func (r *DealRepo) Upsert(ctx context.Context, d models.Deal) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO deals (
			external_id, title, description, price, original_price,
			image_url, deal_url, merchant, temperature, category,
			scraped_at, updated_at, is_expired
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			NOW(), NOW(), FALSE
		)
		ON CONFLICT (external_id) DO UPDATE SET
			title          = EXCLUDED.title,
			description    = EXCLUDED.description,
			price          = EXCLUDED.price,
			original_price = EXCLUDED.original_price,
			image_url      = EXCLUDED.image_url,
			deal_url       = EXCLUDED.deal_url,
			temperature    = EXCLUDED.temperature,
			category       = EXCLUDED.category,
			updated_at     = NOW(),
			is_expired     = FALSE
	`,
		d.ExternalID, d.Title, d.Description, d.Price, d.OriginalPrice,
		d.ImageURL, d.DealURL, d.Merchant, d.Temperature, d.Category,
	)
	if err != nil {
		return fmt.Errorf("upsert deal %s: %w", d.ExternalID, err)
	}
	return nil
}

type ListParams struct {
	Limit  int
	Offset int
}

type ListResult struct {
	Deals   []models.Deal
	Total   int
	HasMore bool
}

func (r *DealRepo) ListPaginated(ctx context.Context, p ListParams) (ListResult, error) {
	if p.Limit <= 0 || p.Limit > 50 {
		p.Limit = 20
	}

	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM deals WHERE is_expired = FALSE`,
	).Scan(&total)
	if err != nil {
		return ListResult{}, fmt.Errorf("count deals: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			d.id, d.external_id, d.title, d.description,
			d.price, d.original_price, d.image_url, d.deal_url,
			d.merchant, d.temperature, d.category,
			d.scraped_at, d.updated_at, d.is_expired,
			COUNT(c.id) AS click_count
		FROM deals d
		LEFT JOIN clicks c ON c.deal_id = d.id
		WHERE d.is_expired = FALSE
		GROUP BY d.id
		ORDER BY d.temperature DESC, d.scraped_at DESC
		LIMIT $1 OFFSET $2
	`, p.Limit, p.Offset)
	if err != nil {
		return ListResult{}, fmt.Errorf("list deals: %w", err)
	}
	defer rows.Close()

	var deals []models.Deal
	for rows.Next() {
		var d models.Deal
		if err := rows.Scan(
			&d.ID, &d.ExternalID, &d.Title, &d.Description,
			&d.Price, &d.OriginalPrice, &d.ImageURL, &d.DealURL,
			&d.Merchant, &d.Temperature, &d.Category,
			&d.ScrapedAt, &d.UpdatedAt, &d.IsExpired,
			&d.ClickCount,
		); err != nil {
			return ListResult{}, fmt.Errorf("scan deal: %w", err)
		}
		deals = append(deals, d)
	}
	if err := rows.Err(); err != nil {
		return ListResult{}, fmt.Errorf("rows error: %w", err)
	}

	return ListResult{
		Deals:   deals,
		Total:   total,
		HasMore: p.Offset+len(deals) < total,
	}, nil
}

func (r *DealRepo) GetByID(ctx context.Context, id int64) (models.Deal, error) {
	var d models.Deal
	err := r.pool.QueryRow(ctx, `
		SELECT
			d.id, d.external_id, d.title, d.description,
			d.price, d.original_price, d.image_url, d.deal_url,
			d.merchant, d.temperature, d.category,
			d.scraped_at, d.updated_at, d.is_expired,
			COUNT(c.id) AS click_count
		FROM deals d
		LEFT JOIN clicks c ON c.deal_id = d.id
		WHERE d.id = $1
		GROUP BY d.id
	`, id).Scan(
		&d.ID, &d.ExternalID, &d.Title, &d.Description,
		&d.Price, &d.OriginalPrice, &d.ImageURL, &d.DealURL,
		&d.Merchant, &d.Temperature, &d.Category,
		&d.ScrapedAt, &d.UpdatedAt, &d.IsExpired,
		&d.ClickCount,
	)
	if err != nil {
		return models.Deal{}, fmt.Errorf("get deal %d: %w", id, err)
	}
	return d, nil
}
