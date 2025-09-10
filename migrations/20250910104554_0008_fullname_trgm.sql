-- +goose Up
-- +goose NO TRANSACTION
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_students_fullname_trgm
    ON students_core USING gin ((last_name || ' ' || first_name || ' ' || COALESCE(middle_name,'')) gin_trgm_ops);

-- +goose Down
-- +goose NO TRANSACTION
DROP INDEX CONCURRENTLY IF EXISTS idx_students_fullname_trgm;
