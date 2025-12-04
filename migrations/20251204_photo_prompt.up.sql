ALTER TABLE bot_configs
    ADD COLUMN photo_style_prompt text NOT NULL DEFAULT 'опиши стиль фото простыми словами, без лишних символов.';

ALTER TABLE classes
    ALTER COLUMN grade TYPE text USING grade::text;