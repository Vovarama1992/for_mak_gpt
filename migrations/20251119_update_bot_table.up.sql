DROP TABLE IF EXISTS bot_prompts;

CREATE TABLE bot_configs (
    bot_id        text PRIMARY KEY,
    token         text NOT NULL,
    model         text NOT NULL,
    style_prompt  text NOT NULL,
    voice_id      text NOT NULL
);