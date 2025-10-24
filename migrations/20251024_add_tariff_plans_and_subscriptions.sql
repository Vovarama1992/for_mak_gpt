CREATE TABLE IF NOT EXISTS tariff_plans (
    id SERIAL PRIMARY KEY,
    code VARCHAR(32) UNIQUE NOT NULL,            -- например: free, pro, family, premium_plus
    name VARCHAR(64) NOT NULL,                   -- человекочитаемое имя
    price NUMERIC(10,2) NOT NULL DEFAULT 0.00,   -- цена в рублях
    period_days INT NOT NULL DEFAULT 30,         -- срок действия в днях
    features JSONB DEFAULT '{}'::jsonb,          -- список возможностей (опционально)
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    bot_id VARCHAR(64) NOT NULL,
    telegram_id BIGINT NOT NULL,
    plan_id INT REFERENCES tariff_plans(id) ON DELETE SET NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'inactive' CHECK (status IN ('inactive', 'active', 'expired', 'pending')),
    started_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (bot_id, telegram_id)
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_bot_telegram ON subscriptions (bot_id, telegram_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions (status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_expires_at ON subscriptions (expires_at);

-- стартовые тарифы
INSERT INTO tariff_plans (code, name, price, period_days, features)
VALUES
('free', 'Бесплатный', 0.00, 30, '{"limits": {"messages_per_day": 3}}'),
('pro', 'Про', 299.00, 30, '{"limits": {"messages_per_day": "unlimited"}, "pdf": true, "reports": true}'),
('family', 'Семейный', 499.00, 30, '{"multi_user": true, "reports": true}'),
('premium_plus', 'Премиум+', 999.00, 30, '{"voice_ai": true, "video": true, "priority": true}')
ON CONFLICT (code) DO NOTHING;