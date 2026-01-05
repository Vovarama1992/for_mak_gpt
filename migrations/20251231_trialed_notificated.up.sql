ALTER TABLE subscriptions
ADD COLUMN trial_notified_at TIMESTAMPTZ;
CREATE INDEX idx_subscriptions_trial_notified
ON subscriptions (trial_notified_at)
WHERE trial_notified_at IS NULL;