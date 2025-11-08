-- +goose Up
CREATE TABLE staff (
                       id SERIAL PRIMARY KEY,
                       full_name TEXT NOT NULL,
                       phone TEXT NOT NULL,
                       position TEXT NOT NULL,
                       education TEXT,
                       category TEXT,
                       ped_experience INT DEFAULT 0,
                       total_experience INT DEFAULT 0,
                       work_start DATE,
                       note TEXT,
                       school_id INT NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
                       created_at TIMESTAMP DEFAULT NOW()
);

-- Индекс для фильтрации
CREATE INDEX idx_staff_school_id ON staff(school_id);

-- +goose Down
DROP TABLE IF EXISTS staff;
