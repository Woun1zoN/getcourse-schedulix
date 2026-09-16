CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT NOT NULL UNIQUE,
    getcourse_cookie TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);