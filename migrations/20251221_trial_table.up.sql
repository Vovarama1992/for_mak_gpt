CREATE TABLE trial_usages (
	bot_id VARCHAR(64) NOT NULL,
	telegram_id BIGINT NOT NULL,
	PRIMARY KEY (bot_id, telegram_id)
);