CREATE TABLE IF NOT EXISTS records (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT NOT NULL,
    bot_id VARCHAR(64) NOT NULL,
    user_ref BIGINT,
    role VARCHAR(16) NOT NULL CHECK (role IN ('user', 'tutor')),
    record_type VARCHAR(16) NOT NULL CHECK (record_type IN ('text', 'image')),
    text_content TEXT,
    image_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_records_bot_telegram_created
    ON records (bot_id, telegram_id, created_at);

CREATE INDEX IF NOT EXISTS idx_records_telegram_created
    ON records (telegram_id, created_at);

CREATE INDEX IF NOT EXISTS idx_records_userref_created
    ON records (user_ref, created_at);
