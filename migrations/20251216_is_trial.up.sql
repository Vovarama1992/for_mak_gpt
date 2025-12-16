ALTER TABLE tariffs
ADD COLUMN is_trial BOOLEAN NOT NULL DEFAULT FALSE;

CREATE UNIQUE INDEX uniq_trial_tariff
ON tariffs (is_trial)
WHERE is_trial = true;