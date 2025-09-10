package models

type Gender string

const (
	GenderMale   Gender = "m"
	GenderFemale Gender = "f"
)

type StudentStatus string

const (
	StatusEnrolled    StudentStatus = "enrolled"
	StatusTransferred StudentStatus = "transferred"
	StatusGraduated   StudentStatus = "graduated"
	StatusExpelled    StudentStatus = "expelled"
)
