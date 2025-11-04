ALTER TABLE subscriptions
ADD COLUMN IF NOT EXISTS yookassa_payment_id VARCHAR(128);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_subscriptions_yookassa_payment_id
    ON subscriptions (yookassa_payment_id)
    WHERE yookassa_payment_id IS NOT NULL;