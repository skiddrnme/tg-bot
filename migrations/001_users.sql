-- github_login хранится в нижнем регистре: GitHub не различает регистр логинов.
CREATE TABLE IF NOT EXISTS users (
    telegram_id  BIGINT PRIMARY KEY,
    github_login TEXT   NOT NULL UNIQUE
);
