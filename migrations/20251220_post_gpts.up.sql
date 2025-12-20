CREATE TABLE text_letter_rules (
    id SERIAL PRIMARY KEY,
    from_char CHAR(1) NOT NULL,
    to_char   CHAR(1) NOT NULL
);

-- words
CREATE TABLE text_word_rules (
    id SERIAL PRIMARY KEY,
    from_word TEXT NOT NULL,
    to_word   TEXT NOT NULL
);