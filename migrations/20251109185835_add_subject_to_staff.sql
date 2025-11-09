-- +goose Up
ALTER TABLE staff ADD COLUMN subject TEXT;

-- +goose Down
ALTER TABLE staff DROP COLUMN IF EXISTS subject;
