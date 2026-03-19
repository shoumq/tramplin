-- +goose Up
ALTER TABLE companies
    ADD COLUMN IF NOT EXISTS avatar_url TEXT NULL,
    ADD COLUMN IF NOT EXISTS avatar_object TEXT NULL;

-- +goose Down
ALTER TABLE companies
    DROP COLUMN IF EXISTS avatar_object,
    DROP COLUMN IF EXISTS avatar_url;
