-- +goose Up
CREATE TABLE manufactures (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL
);

-- +goose StatementBegin
CREATE FUNCTION set_manufactures_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER manufactures_set_updated_at
    BEFORE UPDATE ON manufactures
    FOR EACH ROW
    EXECUTE FUNCTION set_manufactures_updated_at();

-- +goose Down
DROP TABLE manufactures;
DROP FUNCTION set_manufactures_updated_at();
