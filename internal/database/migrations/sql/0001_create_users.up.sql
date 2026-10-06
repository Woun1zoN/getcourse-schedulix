CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT NOT NULL UNIQUE,
    getcourse_cookie TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    getcourse_stream_id BIGINT,
    target_chat_id BIGINT NOT NULL,
    target_chat_title TEXT
);