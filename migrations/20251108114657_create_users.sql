-- +goose Up
CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       email TEXT UNIQUE NOT NULL,
                       password TEXT NOT NULL,
                       role TEXT NOT NULL CHECK (role IN ('roo','school')),
                       created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE schools (
                         id SERIAL PRIMARY KEY,
                         name TEXT NOT NULL,
                         director TEXT,
                         class_count INT DEFAULT 0,
                         student_count INT DEFAULT 0,
                         created_at TIMESTAMP DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS schools;
DROP TABLE IF EXISTS users;
