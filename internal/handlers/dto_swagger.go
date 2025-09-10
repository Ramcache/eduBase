package handlers

import "eduBase/internal/models"

// ===== Requests =====

type CreateStudentCoreRequest struct {
	StudentNumber string  `json:"student_number"`
	LastName      string  `json:"last_name"`
	FirstName     string  `json:"first_name"`
	MiddleName    *string `json:"middle_name"`
	BirthDate     string  `json:"birth_date"` // YYYY-MM-DD
	Gender        string  `json:"gender"`
	Citizenship   *string `json:"citizenship"`
	SchoolID      int     `json:"school_id"`
	ClassLabel    string  `json:"class_label"`
	AdmissionYear int     `json:"admission_year"`
	Status        string  `json:"status"`
	RegAddress    string  `json:"reg_address"`
	FactAddress   string  `json:"fact_address"`
	StudentPhone  *string `json:"student_phone"`
	StudentEmail  *string `json:"student_email"`
}

type UpdateStudentCoreRequest = CreateStudentCoreRequest

type DocumentsUpsertRequest struct {
	StudentID        int     `json:"student_id"`
	SNILS            string  `json:"snils"`
	PassportSeries   *string `json:"passport_series"`
	PassportNumber   *string `json:"passport_number"`
	BirthCertificate *string `json:"birth_certificate"`
	BirthDate        string  `json:"birth_date"`
}

type MedicalUpsertRequest struct {
	StudentID    int     `json:"student_id"`
	Benefits     *string `json:"benefits"`
	MedicalNotes *string `json:"medical_notes"`
	HealthGroup  *int    `json:"health_group"`
	Allergies    *string `json:"allergies"`
	Activities   *string `json:"activities"`
}

type ConsentsUpsertRequest struct {
	StudentID           int     `json:"student_id"`
	ConsentPD           bool    `json:"consent_data_processing"`
	ConsentPDDate       *string `json:"consent_data_processing_date"`
	ConsentPhoto        bool    `json:"consent_photo_publication"`
	ConsentPhotoDate    *string `json:"consent_photo_publication_date"`
	ConsentInternet     bool    `json:"consent_internet_access"`
	ConsentInternetDate *string `json:"consent_internet_access_date"`
}

type ContactAddRequest struct {
	StudentID int    `json:"student_id"`
	FullName  string `json:"full_name"`
	Phone     string `json:"phone"`
	Relation  string `json:"relation"`
}

// ===== Responses =====

type CreateIDResponse struct {
	ID int `json:"id"`
}

type OkResponse struct {
	Status string `json:"status" example:"ok"`
}

type StudentListResponse struct {
	Total  int                      `json:"total"  example:"1"`
	Limit  int                      `json:"limit"  example:"50"`
	Offset int                      `json:"offset" example:"0"`
	Items  []models.StudentListItem `json:"items"`
}
