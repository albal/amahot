CREATE TABLE IF NOT EXISTS clicks (
    id         BIGSERIAL PRIMARY KEY,
    deal_id    BIGINT NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    ip_address INET,
    user_agent TEXT,
    referer    TEXT,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_clicks_deal_id    ON clicks (deal_id);
CREATE INDEX IF NOT EXISTS idx_clicks_clicked_at ON clicks (clicked_at DESC);
