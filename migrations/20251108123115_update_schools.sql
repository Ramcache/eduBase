-- +goose Up
ALTER TABLE schools
    ADD COLUMN user_id INT REFERENCES users(id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE schools DROP COLUMN user_id;
