-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS students_emergency_contacts (
                                                           id BIGSERIAL PRIMARY KEY,
                                                           student_id INT NOT NULL REFERENCES students_core(id) ON DELETE CASCADE,
                                                           full_name VARCHAR(255) NOT NULL,
                                                           phone VARCHAR(30) NOT NULL,
                                                           relation VARCHAR(50) NOT NULL,
                                                           created_at TIMESTAMP NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_emg_by_student ON students_emergency_contacts (student_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS students_emergency_contacts;
-- +goose StatementEnd
