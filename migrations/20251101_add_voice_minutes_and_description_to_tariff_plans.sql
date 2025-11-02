ALTER TABLE tariff_plans
    ADD COLUMN voice_minutes INT NOT NULL DEFAULT 0,
    ADD COLUMN description TEXT;

UPDATE tariff_plans
SET voice_minutes = 0,
    description = 'Текст и фото — бесплатно. Голос недоступен.'
WHERE code = 'free';

UPDATE tariff_plans
SET price = 200.00, voice_minutes = 60,
    description = 'Пробный пакет — 1 час голосового общения.'
WHERE code = 'pro';

UPDATE tariff_plans
SET price = 950.00, voice_minutes = 300,
    description = 'Курс — 5 часов голосовых уроков.'
WHERE code = 'family';

UPDATE tariff_plans
SET price = 1799.00, voice_minutes = 600,
    description = 'Smart — 10 часов голосового общения.'
WHERE code = 'premium_plus';