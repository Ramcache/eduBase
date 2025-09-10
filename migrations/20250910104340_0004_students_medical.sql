-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS students_medical (
                                                student_id INT PRIMARY KEY REFERENCES students_core(id) ON DELETE CASCADE,
                                                benefits TEXT,
                                                medical_notes TEXT,
                                                health_group SMALLINT CHECK (health_group BETWEEN 1 AND 5),
                                                allergies TEXT,
                                                activities TEXT,
                                                updated_at TIMESTAMP NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS students_medical;
-- +goose StatementEnd
