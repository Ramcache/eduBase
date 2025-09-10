-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS students_documents (
                                                  student_id INT PRIMARY KEY REFERENCES students_core(id) ON DELETE CASCADE,
                                                  snils VARCHAR(20) NOT NULL,
                                                  passport_series VARCHAR(10),
                                                  passport_number VARCHAR(20),
                                                  birth_certificate VARCHAR(50),
                                                  updated_at TIMESTAMP NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_docs_snils ON students_documents (snils);
CREATE UNIQUE INDEX IF NOT EXISTS uq_docs_passport ON students_documents (passport_series, passport_number)
    WHERE passport_series IS NOT NULL AND passport_number IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS students_documents;
-- +goose StatementEnd
