package models

import "time"

type Staff struct {
	ID              int        `json:"id"`
	FullName        string     `json:"full_name" validate:"required"`
	Phone           string     `json:"phone" validate:"required"`
	Position        string     `json:"position" validate:"required"`
	Subject         *string    `json:"subject,omitempty"`
	Education       *string    `json:"education,omitempty"`
	Category        *string    `json:"category,omitempty"`
	PedExperience   *int       `json:"ped_experience,omitempty"`
	TotalExperience *int       `json:"total_experience,omitempty"`
	WorkStart       *time.Time `json:"work_start,omitempty"`
	Note            *string    `json:"note,omitempty"`
	SchoolID        int        `json:"school_id"`
	CreatedAt       time.Time  `json:"created_at"`
}
