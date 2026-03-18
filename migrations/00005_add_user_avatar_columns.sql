-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS avatar_url TEXT NULL,
    ADD COLUMN IF NOT EXISTS avatar_object TEXT NULL;

-- +goose Down
ALTER TABLE users
    DROP COLUMN IF EXISTS avatar_object,
    DROP COLUMN IF EXISTS avatar_url;
