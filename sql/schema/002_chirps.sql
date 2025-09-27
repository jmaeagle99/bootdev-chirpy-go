-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto"; -- enables gen_random_uuid()

CREATE TABLE chirps (
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    body TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

CREATE TRIGGER chirps_set_updated_at
BEFORE UPDATE ON chirps
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS chirps_set_updated_at ON chirps;
DROP TABLE IF EXISTS chirps;