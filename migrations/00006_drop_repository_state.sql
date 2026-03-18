-- +goose Up
DROP TABLE IF EXISTS repository_state;

-- +goose Down
CREATE TABLE IF NOT EXISTS repository_state (
    id SMALLINT PRIMARY KEY,
    payload JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
