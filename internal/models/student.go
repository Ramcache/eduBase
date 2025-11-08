package models

import "time"

type Student struct {
	ID        int        `json:"id"`
	FullName  string     `json:"full_name" validate:"required"`
	BirthDate *time.Time `json:"birth_date,omitempty"`
	Gender    *string    `json:"gender,omitempty"`
	Phone     *string    `json:"phone,omitempty"`
	Address   *string    `json:"address,omitempty"`
	Note      *string    `json:"note,omitempty"`
	ClassID   int        `json:"class_id"`
	SchoolID  int        `json:"school_id"`
	CreatedAt time.Time  `json:"created_at"`
}
