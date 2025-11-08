package models

import "time"

type School struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Director     string    `json:"director"`
	ClassCount   int       `json:"class_count"`
	StudentCount int       `json:"student_count"`
	CreatedAt    time.Time `json:"created_at"`
}
