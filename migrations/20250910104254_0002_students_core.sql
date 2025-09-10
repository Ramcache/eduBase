-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS students_core (
                                             id SERIAL PRIMARY KEY,
                                             student_number VARCHAR(50) NOT NULL,

                                             last_name VARCHAR(120) NOT NULL,
                                             first_name VARCHAR(120) NOT NULL,
                                             middle_name VARCHAR(120),

                                             birth_date DATE NOT NULL,
                                             gender gender NOT NULL,
                                             citizenship VARCHAR(100),

                                             school_id INT NOT NULL,
                                             class_label VARCHAR(10) NOT NULL,
                                             admission_year SMALLINT NOT NULL CHECK (admission_year BETWEEN 1990 AND EXTRACT(YEAR FROM now())::INT + 1),
                                             status student_status NOT NULL DEFAULT 'enrolled',

                                             reg_address VARCHAR(512) NOT NULL,
                                             fact_address VARCHAR(512) NOT NULL,
                                             student_phone VARCHAR(30),
                                             student_email VARCHAR(255),

                                             created_at TIMESTAMP NOT NULL DEFAULT now(),
                                             updated_at TIMESTAMP NOT NULL DEFAULT now(),
                                             deleted_at TIMESTAMP,
                                             created_by INT NOT NULL,
                                             updated_by INT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS students_core;
-- +goose StatementEnd
