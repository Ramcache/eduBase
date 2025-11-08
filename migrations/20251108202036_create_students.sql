-- +goose Up
CREATE TABLE students (
                          id SERIAL PRIMARY KEY,
                          full_name TEXT NOT NULL,
                          birth_date DATE,
                          gender TEXT CHECK (gender IN ('male','female')) DEFAULT 'male',
                          phone TEXT,
                          address TEXT,
                          note TEXT,
                          class_id INT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
                          school_id INT NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
                          created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_students_class_id ON students(class_id);
CREATE INDEX idx_students_school_id ON students(school_id);

-- +goose Down
DROP TABLE IF EXISTS students;
