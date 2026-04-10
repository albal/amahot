CREATE TABLE IF NOT EXISTS schema_migrations (
    filename   TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS deals (
    id             BIGSERIAL PRIMARY KEY,
    external_id    TEXT NOT NULL UNIQUE,
    title          TEXT NOT NULL,
    description    TEXT,
    price          TEXT,
    original_price TEXT,
    image_url      TEXT,
    deal_url       TEXT NOT NULL,
    merchant       TEXT NOT NULL DEFAULT 'Amazon',
    temperature    INTEGER NOT NULL DEFAULT 0,
    category       TEXT,
    scraped_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_expired     BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_deals_temperature ON deals (temperature DESC);
CREATE INDEX IF NOT EXISTS idx_deals_external_id ON deals (external_id);
CREATE INDEX IF NOT EXISTS idx_deals_active ON deals (is_expired, temperature DESC);
