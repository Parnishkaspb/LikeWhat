-- +goose Up
ALTER TABLE users ADD CONSTRAINT users_telegram_id_unique UNIQUE (telegram_id);

-- +goose Down
ALTER TABLE users DROP CONSTRAINT users_telegram_id_unique;
