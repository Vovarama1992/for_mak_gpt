ALTER TABLE tariff_plans
ADD COLUMN is_trial BOOLEAN NOT NULL DEFAULT FALSE;

CREATE UNIQUE INDEX uniq_trial_tariff
ON tariff_plans (is_trial)
WHERE is_trial = true;