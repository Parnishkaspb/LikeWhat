-- +goose Up
CREATE TABLE recipes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    BIGINT NOT NULL REFERENCES users (id),
    title      TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_recipes_user_id ON recipes (user_id);

CREATE TABLE recipe_tobaccos (
    recipe_id  UUID NOT NULL REFERENCES recipes (id) ON DELETE CASCADE,
    tobacco_id UUID NOT NULL REFERENCES tobaccos (id),
    percent    NUMERIC(5,2) NOT NULL CHECK (percent > 0 AND percent <= 100),
    PRIMARY KEY (recipe_id, tobacco_id)
);

CREATE INDEX idx_recipe_tobaccos_tobacco_id ON recipe_tobaccos (tobacco_id);

CREATE TABLE recipe_steps (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipe_id   UUID NOT NULL REFERENCES recipes (id) ON DELETE CASCADE,
    step_number INT NOT NULL CHECK (step_number > 0),
    tobacco_id  UUID REFERENCES tobaccos (id),
    what_do     TEXT NOT NULL,
    UNIQUE (recipe_id, step_number)
);

-- +goose StatementBegin
CREATE FUNCTION set_recipes_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER recipes_set_updated_at
    BEFORE UPDATE ON recipes
    FOR EACH ROW
    EXECUTE FUNCTION set_recipes_updated_at();

-- +goose Down
DROP TABLE recipe_steps;
DROP TABLE recipe_tobaccos;
DROP TABLE recipes;
DROP FUNCTION set_recipes_updated_at();
