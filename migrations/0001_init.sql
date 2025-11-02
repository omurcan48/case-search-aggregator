CREATE TABLE IF NOT EXISTS contents (
    id BIGSERIAL PRIMARY KEY,
    external_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    title TEXT NOT NULL,
    type TEXT NOT NULL,
    views INT,
    likes INT,
    reactions INT,
    reading_time_min INT,
    published_at TIMESTAMPTZ,
    score_popularity DOUBLE PRECISION,
    score_relevance DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (external_id, provider)
);

CREATE INDEX IF NOT EXISTS idx_contents_type ON contents(type);
CREATE INDEX IF NOT EXISTS idx_contents_published ON contents(published_at);
