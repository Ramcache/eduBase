-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_core_school_class ON students_core (school_id, class_label);
CREATE INDEX IF NOT EXISTS idx_core_status ON students_core (status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_core_status;
DROP INDEX IF EXISTS idx_core_school_class;
-- +goose StatementEnd
