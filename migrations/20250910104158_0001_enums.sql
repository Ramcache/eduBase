-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN CREATE TYPE gender AS ENUM ('m','f'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
DO $$ BEGIN CREATE TYPE student_status AS ENUM ('enrolled','transferred','graduated','expelled'); EXCEPTION WHEN duplicate_object THEN NULL; END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS student_status;
DROP TYPE IF EXISTS gender;
-- +goose StatementEnd
