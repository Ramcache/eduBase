-- +goose Up
CREATE TABLE classes (
                         id SERIAL PRIMARY KEY,
                         name TEXT NOT NULL,
                         grade INT NOT NULL,
                         school_id INT NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
                         student_count INT DEFAULT 0,
                         created_at TIMESTAMP DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS classes;
