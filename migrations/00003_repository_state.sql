-- +goose Up
CREATE TABLE IF NOT EXISTS repository_state (
    id SMALLINT PRIMARY KEY,
    payload JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS repository_state;
