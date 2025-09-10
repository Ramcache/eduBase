package models

import "time"

// Core (А+Б+В)
type StudentCore struct {
	ID            int       `json:"id"`
	StudentNumber string    `json:"student_number"`
	LastName      string    `json:"last_name"`
	FirstName     string    `json:"first_name"`
	MiddleName    *string   `json:"middle_name,omitempty"`
	BirthDate     time.Time `json:"birth_date"`
	Gender        Gender    `json:"gender"`
	Citizenship   *string   `json:"citizenship,omitempty"`

	SchoolID      int           `json:"school_id"`
	ClassLabel    string        `json:"class_label"`
	AdmissionYear int           `json:"admission_year"`
	Status        StudentStatus `json:"status"`

	RegAddress   string  `json:"reg_address"`
	FactAddress  string  `json:"fact_address"`
	StudentPhone *string `json:"student_phone,omitempty"`
	StudentEmail *string `json:"student_email,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedBy int        `json:"created_by"`
	UpdatedBy *int       `json:"updated_by,omitempty"`
}

// Documents (Г)
type StudentDocuments struct {
	StudentID        int       `json:"student_id"`
	SNILS            string    `json:"snils"`
	PassportSeries   *string   `json:"passport_series,omitempty"`
	PassportNumber   *string   `json:"passport_number,omitempty"`
	BirthCertificate *string   `json:"birth_certificate,omitempty"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Medical (Д)
type StudentMedical struct {
	StudentID    int       `json:"student_id"`
	Benefits     *string   `json:"benefits,omitempty"`
	MedicalNotes *string   `json:"medical_notes,omitempty"`
	HealthGroup  *int      `json:"health_group,omitempty"`
	Allergies    *string   `json:"allergies,omitempty"`
	Activities   *string   `json:"activities,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Consents (Ж)
type StudentConsents struct {
	StudentID           int        `json:"student_id"`
	ConsentPD           bool       `json:"consent_data_processing"`
	ConsentPDDate       *time.Time `json:"consent_data_processing_date,omitempty"`
	ConsentPhoto        bool       `json:"consent_photo_publication"`
	ConsentPhotoDate    *time.Time `json:"consent_photo_publication_date,omitempty"`
	ConsentInternet     bool       `json:"consent_internet_access"`
	ConsentInternetDate *time.Time `json:"consent_internet_access_date,omitempty"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// Emergency contacts
type EmergencyContact struct {
	ID        int       `json:"id"`
	StudentID int       `json:"student_id"`
	FullName  string    `json:"full_name"`
	Phone     string    `json:"phone"`
	Relation  string    `json:"relation"`
	CreatedAt time.Time `json:"created_at"`
}

// Aggregate view
type StudentView struct {
	Core     StudentCore        `json:"core"`
	Docs     *StudentDocuments  `json:"documents,omitempty"`
	Medical  *StudentMedical    `json:"medical,omitempty"`
	Consents *StudentConsents   `json:"consents,omitempty"`
	Contacts []EmergencyContact `json:"contacts,omitempty"`
}

// For list + export
type StudentListItem struct {
	ID            int           `json:"id"`
	StudentNumber string        `json:"student_number"`
	FullName      string        `json:"full_name"`
	BirthDate     time.Time     `json:"birth_date"`
	Gender        Gender        `json:"gender"`
	SchoolID      int           `json:"school_id"`
	ClassLabel    string        `json:"class_label"`
	AdmissionYear int           `json:"admission_year"`
	Status        StudentStatus `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
}
