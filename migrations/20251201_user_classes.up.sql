ALTER TABLE bot_configs
    DROP COLUMN style_prompt;

ALTER TABLE bot_configs
    ADD COLUMN text_style_prompt  text NOT NULL DEFAULT 'используй любые символы.',
    ADD COLUMN voice_style_prompt text NOT NULL DEFAULT 'используй только словаю, никаких символов даже точек и запятых.';

-- 2) таблица class_prompts (справочник)
CREATE TABLE classes (
    id    serial PRIMARY KEY,
    grade int UNIQUE NOT NULL  -- 1..11
);

-------------------------------------------------------
-- 3) новая таблица class_prompts (промпты класса)
-------------------------------------------------------
CREATE TABLE class_prompts (
    id        serial PRIMARY KEY,
    class_id  int NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    prompt    text NOT NULL
);

-------------------------------------------------------
-- 4) новая таблица user_classes (связка юзер → класс)
-------------------------------------------------------
CREATE TABLE user_classes (
    bot_id       text   NOT NULL,
    telegram_id  bigint NOT NULL,
    class_id     int    NOT NULL REFERENCES classes(id) ON DELETE RESTRICT,
    PRIMARY KEY (bot_id, telegram_id)
);