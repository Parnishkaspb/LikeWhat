-- +goose Up
CREATE TABLE tobaccos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    taste TEXT NOT NULL,
    photo TEXT,
    manufacture_id UUID NOT NULL REFERENCES manufactures (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_tobaccos_manufacture_id ON tobaccos (manufacture_id);

-- +goose StatementBegin
CREATE FUNCTION set_tobaccos_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER tobaccos_set_updated_at
    BEFORE UPDATE ON tobaccos
    FOR EACH ROW
    EXECUTE FUNCTION set_tobaccos_updated_at();

-- +goose Down
DROP TABLE tobaccos;
DROP FUNCTION set_tobaccos_updated_at();
