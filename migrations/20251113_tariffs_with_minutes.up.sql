DELETE FROM tariff_plans;

-- 2) Удаляем поле, которое больше не нужно
ALTER TABLE tariff_plans
    DROP COLUMN IF EXISTS period_days;

-- 3) Добавляем duration_minutes — единую ось времени
ALTER TABLE tariff_plans
    ADD COLUMN IF NOT EXISTS duration_minutes INT NOT NULL DEFAULT 0;

-- 4) Вставляем новые тарифы

INSERT INTO tariff_plans (
    code,
    name,
    price,
    duration_minutes,
    voice_minutes,
    description
)
VALUES
(
    'hour300',
    '1 час обучения',
    300,
    60,     -- доступ 60 минут
    10,     -- голос 10 минут
    '60 минут доступа + 10 минут голоса'
),
(
    'daily690',
    'Учусь каждый день',
    690,
    7 * 24 * 60,     -- 7 дней в минутах
    25,
    '7 дней доступа + 25 минут голоса'
),
(
    'max990',
    'Максимум',
    990,
    5 * 24 * 60,     -- 5 дней в минутах
    40,
    '5 дней доступа + 40 минут голоса'
);