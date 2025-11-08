package models

import "time"

type Class struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Grade        int       `json:"grade"`
	SchoolID     int       `json:"school_id"`
	StudentCount int       `json:"student_count"`
	CreatedAt    time.Time `json:"created_at"`
}
