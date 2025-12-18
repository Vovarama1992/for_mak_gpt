CREATE TABLE bot_auth (
    id SERIAL PRIMARY KEY,
    password TEXT NOT NULL
);

-- начальный пароль
INSERT INTO bot_auth (password) VALUES ('swarms_save_theWorld');