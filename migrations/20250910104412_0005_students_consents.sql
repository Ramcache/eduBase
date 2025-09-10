-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS students_consents (
                                                 student_id INT PRIMARY KEY REFERENCES students_core(id) ON DELETE CASCADE,
                                                 consent_data_processing BOOLEAN NOT NULL DEFAULT FALSE,
                                                 consent_data_processing_date DATE,
                                                 consent_photo_publication BOOLEAN NOT NULL DEFAULT FALSE,
                                                 consent_photo_publication_date DATE,
                                                 consent_internet_access BOOLEAN NOT NULL DEFAULT FALSE,
                                                 consent_internet_access_date DATE,
                                                 updated_at TIMESTAMP NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS students_consents;
-- +goose StatementEnd
