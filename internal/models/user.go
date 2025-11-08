package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `json:"role"` // roo | school
	CreatedAt time.Time `json:"created_at"`
}
