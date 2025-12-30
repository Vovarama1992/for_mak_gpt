BEGIN;

--------------------------------------------------
-- 1. CLASSES
--------------------------------------------------
ALTER TABLE classes
ADD COLUMN bot_id text;

UPDATE classes
SET bot_id = (SELECT bot_id FROM bot_configs LIMIT 1);

ALTER TABLE classes
ALTER COLUMN bot_id SET NOT NULL;

ALTER TABLE classes
ADD CONSTRAINT fk_classes_bot
FOREIGN KEY (bot_id) REFERENCES bot_configs(bot_id)
ON DELETE CASCADE;

DROP INDEX IF EXISTS classes_grade_key;
CREATE UNIQUE INDEX classes_bot_grade_uq
ON classes (bot_id, grade);


--------------------------------------------------
-- 2. USER_CLASSES
--------------------------------------------------
ALTER TABLE user_classes
ADD COLUMN bot_id text;

UPDATE user_classes
SET bot_id = (SELECT bot_id FROM bot_configs LIMIT 1);

ALTER TABLE user_classes
ALTER COLUMN bot_id SET NOT NULL;

ALTER TABLE user_classes
DROP CONSTRAINT user_classes_pkey;

ALTER TABLE user_classes
ADD CONSTRAINT user_classes_pkey
PRIMARY KEY (bot_id, telegram_id);

ALTER TABLE user_classes
ADD CONSTRAINT fk_user_classes_bot
FOREIGN KEY (bot_id) REFERENCES bot_configs(bot_id)
ON DELETE CASCADE;


--------------------------------------------------
-- 3. CLASS_PROMPTS
--------------------------------------------------
ALTER TABLE class_prompts
ADD COLUMN bot_id text;

UPDATE class_prompts
SET bot_id = (SELECT bot_id FROM bot_configs LIMIT 1);

ALTER TABLE class_prompts
ALTER COLUMN bot_id SET NOT NULL;

ALTER TABLE class_prompts
ADD CONSTRAINT fk_class_prompts_bot
FOREIGN KEY (bot_id) REFERENCES bot_configs(bot_id)
ON DELETE CASCADE;


--------------------------------------------------
-- 4. TARIFF_PLANS
--------------------------------------------------
ALTER TABLE tariff_plans
ADD COLUMN bot_id text;

UPDATE tariff_plans
SET bot_id = (SELECT bot_id FROM bot_configs LIMIT 1);

ALTER TABLE tariff_plans
ALTER COLUMN bot_id SET NOT NULL;

ALTER TABLE tariff_plans
ADD CONSTRAINT fk_tariff_plans_bot
FOREIGN KEY (bot_id) REFERENCES bot_configs(bot_id)
ON DELETE CASCADE;

DROP INDEX IF EXISTS tariff_plans_code_key;
CREATE UNIQUE INDEX tariff_plans_bot_code_uq
ON tariff_plans (bot_id, code);

DROP INDEX IF EXISTS uniq_trial_tariff;
CREATE UNIQUE INDEX uniq_trial_tariff_per_bot
ON tariff_plans (bot_id)
WHERE is_trial = true;


--------------------------------------------------
-- 5. MINUTE_PACKAGES
--------------------------------------------------
ALTER TABLE minute_packages
ADD COLUMN bot_id text;

UPDATE minute_packages
SET bot_id = (SELECT bot_id FROM bot_configs LIMIT 1);

ALTER TABLE minute_packages
ALTER COLUMN bot_id SET NOT NULL;

ALTER TABLE minute_packages
ADD CONSTRAINT fk_minute_packages_bot
FOREIGN KEY (bot_id) REFERENCES bot_configs(bot_id)
ON DELETE CASCADE;

DROP INDEX IF EXISTS minute_packages_name_key;
CREATE UNIQUE INDEX minute_packages_bot_name_uq
ON minute_packages (bot_id, name);

COMMIT;