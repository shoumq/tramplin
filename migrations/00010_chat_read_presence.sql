-- +goose Up
ALTER TABLE chat_messages
    ADD COLUMN read_at TIMESTAMPTZ NULL;

CREATE INDEX idx_chat_messages_unread
    ON chat_messages(conversation_id, read_at)
    WHERE read_at IS NULL;

CREATE TABLE user_presence (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    is_online BOOLEAN NOT NULL DEFAULT FALSE,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_presence_online ON user_presence(is_online, last_seen_at DESC);

-- +goose Down
DROP TABLE IF EXISTS user_presence;

DROP INDEX IF EXISTS idx_chat_messages_unread;

ALTER TABLE chat_messages
    DROP COLUMN IF EXISTS read_at;
