package services

import (
	"context"
	"errors"

	"eduBase/internal/models"
	"eduBase/internal/repository"
	"github.com/jackc/pgx/v5"
)

type StaffService struct {
	repo *repository.StaffRepository
	db   *pgx.Conn
}

func NewStaffService(repo *repository.StaffRepository) *StaffService {
	return &StaffService{repo: repo, db: repo.DB()}
}

func (s *StaffService) RepoDB() *pgx.Conn {
	return s.db
}

func (s *StaffService) schoolRepo() *repository.SchoolRepository {
	return repository.NewSchoolRepository(s.db)
}

func (s *StaffService) Create(ctx context.Context, staff *models.Staff, role string, userID int) error {
	switch role {
	case "roo":
		// roo — или сам ставит school_id, или можно запретить это, если не нужно
	case "school":
		school, err := s.schoolRepo().GetByUserID(ctx, userID)
		if err != nil {
			return err
		}
		// жёстко привязываем сотрудника к школе пользователя
		staff.SchoolID = school.ID
	default:
		return errors.New("access denied")
	}

	return s.repo.Create(ctx, staff)
}

// GetAll — сам определяет schoolID по роли и userID + фильтры + пагинация
func (s *StaffService) GetAll(ctx context.Context, role string, userID int, f repository.StaffFilter) ([]models.Staff, error) {
	var schoolID *int

	switch role {
	case "roo":
		// видит всех
	case "school":
		school, err := s.schoolRepo().GetByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		schoolID = &school.ID
	default:
		return nil, errors.New("access denied")
	}

	// дефолтная пагинация
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	return s.repo.GetAll(ctx, schoolID, f)
}

func (s *StaffService) GetByID(ctx context.Context, id int, role string, userID int) (*models.Staff, error) {
	staff, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role == "school" {
		school, err := s.schoolRepo().GetByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if staff.SchoolID != school.ID {
			return nil, errors.New("access denied")
		}
	}

	return staff, nil
}

// Update — защита от подмены school_id, проверка доступа внутри сервиса
func (s *StaffService) Update(ctx context.Context, id int, staff *models.Staff, role string, userID int) (bool, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrStaffNotFound) {
			return false, nil
		}
		return false, err
	}

	switch role {
	case "roo":
		// roo может менять всё, но school_id мы тут намеренно НЕ обновляем
		// если нужно разрешить ему переносить сотрудника между школами —
		// это делается отдельным методом/логикой
	case "school":
		school, err := s.schoolRepo().GetByUserID(ctx, userID)
		if err != nil {
			return false, err
		}
		if existing.SchoolID != school.ID {
			// пытается редактировать чужого сотрудника
			return false, errors.New("access denied")
		}
		// фиксируем school_id в структуре, независимо от того,
		// что пришло в payload (подменить не получится)
		staff.SchoolID = school.ID
	default:
		return false, errors.New("access denied")
	}

	rows, err := s.repo.Update(ctx, id, staff)
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

// Delete — roo может удалить любого, school — только своего
func (s *StaffService) Delete(ctx context.Context, id int, role string, userID int) (bool, error) {
	var schoolID *int

	switch role {
	case "roo":
		// без ограничения
	case "school":
		school, err := s.schoolRepo().GetByUserID(ctx, userID)
		if err != nil {
			return false, err
		}
		schoolID = &school.ID
	default:
		return false, errors.New("access denied")
	}

	rows, err := s.repo.Delete(ctx, id, schoolID)
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (s *StaffService) GetStats(ctx context.Context, role string) (map[string]int, error) {
	if role != "roo" {
		return nil, errors.New("access denied")
	}
	return s.repo.GetStats(ctx)
}
