ALTER TABLE bot_configs
    DROP COLUMN style_prompt;

ALTER TABLE bot_configs
    ADD COLUMN text_style_prompt  text NOT NULL DEFAULT 'используй любые символы.',
    ADD COLUMN voice_style_prompt text NOT NULL DEFAULT 'используй только словаю, никаких символов даже точек и запятых.';

-- 2) таблица class_prompts (справочник)
CREATE TABLE class_prompts (
    id      serial PRIMARY KEY,
    class   int UNIQUE NOT NULL,     -- 1..11
    prompt  text NOT NULL
);

-- 3) таблица user_classes
CREATE TABLE user_classes (
    bot_id       text NOT NULL,
    telegram_id  bigint NOT NULL,
    class_id     int NOT NULL REFERENCES class_prompts(id) ON DELETE RESTRICT,
    PRIMARY KEY (bot_id, telegram_id)
);