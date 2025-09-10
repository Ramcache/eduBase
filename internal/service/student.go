package service

import (
	"context"
	"errors"
	"time"

	"eduBase/internal/models"
	"eduBase/internal/repository"
	"eduBase/internal/validators"
)

type StudentService interface {
	CreateCore(ctx context.Context, dto CreateCoreDTO, createdBy int) (int, error)
	UpdateCore(ctx context.Context, id int, dto UpdateCoreDTO, updatedBy int) error
	GetAggregate(ctx context.Context, id int) (*models.StudentView, error)
	List(ctx context.Context, f repository.StudentFilters, limit, offset int) ([]models.StudentListItem, int, error)

	UpsertDocuments(ctx context.Context, dto DocumentsDTO) error
	UpsertMedical(ctx context.Context, dto MedicalDTO) error
	UpsertConsents(ctx context.Context, dto ConsentsDTO) error

	AddContact(ctx context.Context, dto ContactDTO) (int, error)
	DeleteContact(ctx context.Context, id int) error
	CollectExportData(ctx context.Context, f repository.StudentFilters, limit int) (
		students []models.StudentCore,
		docs map[int]*models.StudentDocuments,
		medical map[int]*models.StudentMedical,
		consents map[int]*models.StudentConsents,
		contacts map[int][]models.EmergencyContact,
		total int, err error,
	)
	Import(ctx context.Context,
		rows []ImportStudent,
		contacts []ImportContact,
		operatorID int,
		replaceContacts bool,
	) ([]ImportResult, error)
}

type studentService struct {
	coreRepo     repository.StudentCoreRepository
	docsRepo     repository.DocumentsRepository
	medRepo      repository.MedicalRepository
	conRepo      repository.ConsentsRepository
	contactsRepo repository.ContactsRepository
}

func NewStudentService(
	core repository.StudentCoreRepository,
	docs repository.DocumentsRepository,
	med repository.MedicalRepository,
	con repository.ConsentsRepository,
	contacts repository.ContactsRepository,
) StudentService {
	return &studentService{
		coreRepo:     core,
		docsRepo:     docs,
		medRepo:      med,
		conRepo:      con,
		contactsRepo: contacts,
	}
}

// DTOs

type CreateCoreDTO struct {
	StudentNumber string
	LastName      string
	FirstName     string
	MiddleName    *string
	BirthDate     string
	Gender        string
	Citizenship   *string
	SchoolID      int
	ClassLabel    string
	AdmissionYear int
	Status        string
	RegAddress    string
	FactAddress   string
	StudentPhone  *string
	StudentEmail  *string
}

type UpdateCoreDTO = CreateCoreDTO

func (s *studentService) CreateCore(ctx context.Context, dto CreateCoreDTO, createdBy int) (int, error) {
	if err := validators.BasicCore(dto.LastName, dto.FirstName, dto.BirthDate, dto.Gender, dto.SchoolID, dto.ClassLabel, dto.AdmissionYear, dto.RegAddress, dto.FactAddress, dto.Status); err != nil {
		return 0, err
	}
	bd, err := time.Parse("2006-01-02", dto.BirthDate)
	if err != nil {
		return 0, errors.New("birth_date must be YYYY-MM-DD")
	}

	m := &models.StudentCore{
		StudentNumber: dto.StudentNumber,
		LastName:      dto.LastName,
		FirstName:     dto.FirstName,
		MiddleName:    dto.MiddleName,
		BirthDate:     bd,
		Gender:        models.Gender(dto.Gender),
		Citizenship:   dto.Citizenship,
		SchoolID:      dto.SchoolID,
		ClassLabel:    dto.ClassLabel,
		AdmissionYear: dto.AdmissionYear,
		Status:        models.StudentStatus(dto.Status),
		RegAddress:    dto.RegAddress,
		FactAddress:   dto.FactAddress,
		StudentPhone:  dto.StudentPhone,
		StudentEmail:  dto.StudentEmail,
		CreatedBy:     createdBy,
	}
	return s.coreRepo.Create(ctx, m)
}

func (s *studentService) UpdateCore(ctx context.Context, id int, dto UpdateCoreDTO, updatedBy int) error {
	if err := validators.BasicCore(dto.LastName, dto.FirstName, dto.BirthDate, dto.Gender, dto.SchoolID, dto.ClassLabel, dto.AdmissionYear, dto.RegAddress, dto.FactAddress, dto.Status); err != nil {
		return err
	}
	bd, err := time.Parse("2006-01-02", dto.BirthDate)
	if err != nil {
		return errors.New("birth_date must be YYYY-MM-DD")
	}
	m := &models.StudentCore{
		ID:            id,
		StudentNumber: dto.StudentNumber,
		LastName:      dto.LastName,
		FirstName:     dto.FirstName,
		MiddleName:    dto.MiddleName,
		BirthDate:     bd,
		Gender:        models.Gender(dto.Gender),
		Citizenship:   dto.Citizenship,
		SchoolID:      dto.SchoolID,
		ClassLabel:    dto.ClassLabel,
		AdmissionYear: dto.AdmissionYear,
		Status:        models.StudentStatus(dto.Status),
		RegAddress:    dto.RegAddress,
		FactAddress:   dto.FactAddress,
		StudentPhone:  dto.StudentPhone,
		StudentEmail:  dto.StudentEmail,
		UpdatedBy:     &updatedBy,
	}
	return s.coreRepo.Update(ctx, m)
}

func (s *studentService) List(ctx context.Context, f repository.StudentFilters, limit, offset int) ([]models.StudentListItem, int, error) {
	return s.coreRepo.List(ctx, f, limit, offset)
}

func (s *studentService) GetAggregate(ctx context.Context, id int) (*models.StudentView, error) {
	core, err := s.coreRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	docs, err := s.docsRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	medical, err := s.medRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	consents, err := s.conRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	contacts, err := s.contactsRepo.ListByStudent(ctx, id)
	if err != nil {
		return nil, err
	}

	return &models.StudentView{
		Core:     *core,
		Docs:     docs,
		Medical:  medical,
		Consents: consents,
		Contacts: contacts,
	}, nil
}

// Documents

type DocumentsDTO struct {
	StudentID        int
	SNILS            string
	PassportSeries   *string
	PassportNumber   *string
	BirthCertificate *string
	BirthDate        string // for conditional rule
}

func (s *studentService) UpsertDocuments(ctx context.Context, dto DocumentsDTO) error {
	bd, err := time.Parse("2006-01-02", dto.BirthDate)
	if err != nil {
		return errors.New("birth_date must be YYYY-MM-DD")
	}
	sn := &dto.SNILS
	if err := validators.CheckDocs(bd, sn, dto.PassportSeries, dto.PassportNumber, dto.BirthCertificate); err != nil {
		return err
	}
	return s.docsRepo.Upsert(ctx, dto.StudentID, dto.SNILS, dto.PassportSeries, dto.PassportNumber, dto.BirthCertificate)
}

// Medical

type MedicalDTO struct {
	StudentID    int
	Benefits     *string
	MedicalNotes *string
	HealthGroup  *int
	Allergies    *string
	Activities   *string
}

func (s *studentService) UpsertMedical(ctx context.Context, dto MedicalDTO) error {
	return s.medRepo.Upsert(ctx, dto.StudentID, dto.Benefits, dto.MedicalNotes, dto.HealthGroup, dto.Allergies, dto.Activities)
}

// Consents

type ConsentsDTO struct {
	StudentID           int
	ConsentPD           bool
	ConsentPDDate       *string
	ConsentPhoto        bool
	ConsentPhotoDate    *string
	ConsentInternet     bool
	ConsentInternetDate *string
}

func (s *studentService) UpsertConsents(ctx context.Context, dto ConsentsDTO) error {
	pd, err := validators.ParseYMD(dto.ConsentPDDate)
	if err != nil {
		return err
	}
	ph, err := validators.ParseYMD(dto.ConsentPhotoDate)
	if err != nil {
		return err
	}
	net, err := validators.ParseYMD(dto.ConsentInternetDate)
	if err != nil {
		return err
	}
	return s.conRepo.Upsert(ctx, dto.StudentID, dto.ConsentPD, pd, dto.ConsentPhoto, ph, dto.ConsentInternet, net)
}

// Contacts

type ContactDTO struct {
	StudentID int
	FullName  string
	Phone     string
	Relation  string
}

func (s *studentService) AddContact(ctx context.Context, dto ContactDTO) (int, error) {
	return s.contactsRepo.Add(ctx, dto.StudentID, dto.FullName, dto.Phone, dto.Relation)
}
func (s *studentService) DeleteContact(ctx context.Context, id int) error {
	return s.contactsRepo.Delete(ctx, id)
}

func (s *studentService) CollectExportData(ctx context.Context, f repository.StudentFilters, limit int) (
	[]models.StudentCore,
	map[int]*models.StudentDocuments,
	map[int]*models.StudentMedical,
	map[int]*models.StudentConsents,
	map[int][]models.EmergencyContact,
	int, error,
) {
	if limit <= 0 || limit > 100000 {
		limit = 50000
	}
	cores, total, err := s.coreRepo.ListFull(ctx, f, limit, 0)
	if err != nil {
		return nil, nil, nil, nil, nil, 0, err
	}

	ids := make([]int, 0, len(cores))
	for _, c := range cores {
		ids = append(ids, c.ID)
	}

	dmap, err := s.docsRepo.BulkByStudentIDs(ctx, ids)
	if err != nil {
		return nil, nil, nil, nil, nil, 0, err
	}
	mmap, err := s.medRepo.BulkByStudentIDs(ctx, ids)
	if err != nil {
		return nil, nil, nil, nil, nil, 0, err
	}
	cmap, err := s.conRepo.BulkByStudentIDs(ctx, ids)
	if err != nil {
		return nil, nil, nil, nil, nil, 0, err
	}
	emap, err := s.contactsRepo.BulkByStudentIDs(ctx, ids)
	if err != nil {
		return nil, nil, nil, nil, nil, 0, err
	}

	return cores, dmap, mmap, cmap, emap, total, nil
}
