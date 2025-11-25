CREATE TABLE minute_packages (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,        -- например "100", "300", "1000"
    minutes INTEGER NOT NULL,         -- 100 / 300 / 1000 минут
    price NUMERIC(10,2) NOT NULL,     -- цена в рублях
    active BOOLEAN NOT NULL DEFAULT true
);