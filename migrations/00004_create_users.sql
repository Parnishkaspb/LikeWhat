-- +goose Up
CREATE TABLE users (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    telegram_id BIGINT NOT NULL,
    nick_name   TEXT NOT NULL,
    name        TEXT NOT NULL
);

-- +goose Down
DROP TABLE users;
