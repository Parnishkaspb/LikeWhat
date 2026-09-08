-- +goose Up
ALTER TABLE manufactures ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- +goose Down
ALTER TABLE manufactures ALTER COLUMN id DROP DEFAULT;
