CREATE TABLE IF NOT EXISTS urls (
    code         CHAR(8) PRIMARY KEY,
    original_url TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
