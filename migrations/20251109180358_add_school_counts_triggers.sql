-- +goose Up
BEGIN;

-- === Функция пересчёта количества классов ===
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_school_class_count()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        UPDATE schools
        SET class_count = (
            SELECT COUNT(*) FROM classes WHERE school_id = OLD.school_id
        )
        WHERE id = OLD.school_id;
        RETURN OLD;
    ELSE
        UPDATE schools
        SET class_count = (
            SELECT COUNT(*) FROM classes WHERE school_id = NEW.school_id
        )
        WHERE id = NEW.school_id;

        IF TG_OP = 'UPDATE' AND NEW.school_id IS DISTINCT FROM OLD.school_id THEN
            UPDATE schools
            SET class_count = (
                SELECT COUNT(*) FROM classes WHERE school_id = OLD.school_id
            )
            WHERE id = OLD.school_id;
        END IF;

        RETURN NEW;
    END IF;
END;
$$;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_update_school_class_count ON classes;
CREATE TRIGGER trg_update_school_class_count
    AFTER INSERT OR DELETE OR UPDATE OF school_id
    ON classes
    FOR EACH ROW
EXECUTE FUNCTION update_school_class_count();

-- === Функция пересчёта количества учеников ===
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_school_student_count()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        UPDATE schools
        SET student_count = (
            SELECT COUNT(*) FROM students WHERE school_id = OLD.school_id
        )
        WHERE id = OLD.school_id;
        RETURN OLD;
    ELSE
        UPDATE schools
        SET student_count = (
            SELECT COUNT(*) FROM students WHERE school_id = NEW.school_id
        )
        WHERE id = NEW.school_id;

        IF TG_OP = 'UPDATE' AND NEW.school_id IS DISTINCT FROM OLD.school_id THEN
            UPDATE schools
            SET student_count = (
                SELECT COUNT(*) FROM students WHERE school_id = OLD.school_id
            )
            WHERE id = OLD.school_id;
        END IF;

        RETURN NEW;
    END IF;
END;
$$;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_update_school_student_count ON students;
CREATE TRIGGER trg_update_school_student_count
    AFTER INSERT OR DELETE OR UPDATE OF school_id
    ON students
    FOR EACH ROW
EXECUTE FUNCTION update_school_student_count();

-- === Разовый бэкофилл текущих значений ===
UPDATE schools SET class_count = COALESCE(class_count, 0), student_count = COALESCE(student_count, 0);

WITH cls AS (
    SELECT school_id, COUNT(*) AS cnt FROM classes GROUP BY school_id
)
UPDATE schools s
SET class_count = cls.cnt
FROM cls
WHERE s.id = cls.school_id;

WITH st AS (
    SELECT school_id, COUNT(*) AS cnt FROM students GROUP BY school_id
)
UPDATE schools s
SET student_count = st.cnt
FROM st
WHERE s.id = st.school_id;

COMMIT;

-- +goose Down
BEGIN;
DROP TRIGGER IF EXISTS trg_update_school_class_count ON classes;
DROP FUNCTION IF EXISTS update_school_class_count();
DROP TRIGGER IF EXISTS trg_update_school_student_count ON students;
DROP FUNCTION IF EXISTS update_school_student_count();
COMMIT;
