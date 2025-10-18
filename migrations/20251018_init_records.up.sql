CREATE TABLE IF NOT EXISTS records (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT NOT NULL,                 -- рабочий идентификатор сейчас
    user_ref BIGINT,                             -- будущая ссылка на users.id (пока без FK)
    role VARCHAR(16) NOT NULL CHECK (role IN ('user', 'tutor')),
    record_type VARCHAR(16) NOT NULL CHECK (record_type IN ('text', 'image')),
    text_content TEXT,
    image_url TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- индексы под текущие запросы
CREATE INDEX IF NOT EXISTS idx_records_telegram_created
    ON records (telegram_id, created_at);

-- заранее индекс под будущую связку, чтобы потом не трогать структуру
CREATE INDEX IF NOT EXISTS idx_records_userref_created
    ON records (user_ref, created_at);
