package models

import "time"

type UserInfo struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type School struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Director     string    `json:"director"`
	ClassCount   int       `json:"class_count"`
	StudentCount int       `json:"student_count"`
	UserID       int       `json:"user_id"`
	User         *UserInfo `json:"user,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
