CREATE TABLE history_state (
    bot_id           TEXT   NOT NULL,
    telegram_id      BIGINT NOT NULL,
    last_n_records   INT    NOT NULL DEFAULT 0,  -- сколько последних сообщений брать
    total_tokens     INT    NOT NULL DEFAULT 0,  -- суммарный вес этих сообщений в токенах
    updated_at       TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (bot_id, telegram_id)
);

ALTER TABLE subscriptions
ADD COLUMN voice_minutes REAL NOT NULL DEFAULT 0;