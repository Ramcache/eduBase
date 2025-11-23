package services

import (
	"context"
	"errors"

	"eduBase/internal/models"
	"eduBase/internal/repository"
)

type StudentService struct {
	repo       *repository.StudentRepository
	classRepo  *repository.ClassRepository
	schoolRepo *repository.SchoolRepository
}

func NewStudentService(r *repository.StudentRepository, cr *repository.ClassRepository, sr *repository.SchoolRepository) *StudentService {
	return &StudentService{repo: r, classRepo: cr, schoolRepo: sr}
}

// ==== CRUD с учётом роли и доступа ====

func (s *StudentService) Create(ctx context.Context, st *models.Student, role string, userID int) error {
	switch role {
	case "school":
		school, err := s.schoolRepo.GetByUserID(ctx, userID)
		if err != nil {
			return err
		}
		st.SchoolID = school.ID
	default:
		return errors.New("access denied")
	}

	if err := s.repo.Create(ctx, st); err != nil {
		return err
	}
	return s.UpdateCounts(ctx, st.SchoolID, st.ClassID)
}

func (s *StudentService) GetAll(ctx context.Context, role string, userID int, f repository.StudentFilter) ([]models.Student, error) {
	var schoolID *int

	switch role {
	case "roo":
		// видит всех
	case "school":
		school, err := s.schoolRepo.GetByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		schoolID = &school.ID
	default:
		return nil, errors.New("access denied")
	}

	if f.Offset < 0 {
		f.Offset = 0
	}

	return s.repo.GetAll(ctx, schoolID, f)
}

func (s *StudentService) GetByID(ctx context.Context, id int, role string, userID int) (*models.Student, error) {
	st, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	switch role {
	case "roo":
		// ок
	case "school":
		school, err := s.schoolRepo.GetByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if st.SchoolID != school.ID {
			return nil, errors.New("access denied")
		}
	default:
		return nil, errors.New("access denied")
	}

	return st, nil
}

func (s *StudentService) Update(ctx context.Context, id int, st *models.Student, role string, userID int) (bool, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrStudentNotFound) {
			return false, nil
		}
		return false, err
	}

	switch role {
	case "roo":
		// может менять любого ученика (кроме school_id)
	case "school":
		school, err := s.schoolRepo.GetByUserID(ctx, userID)
		if err != nil {
			return false, err
		}
		if existing.SchoolID != school.ID {
			return false, errors.New("access denied")
		}
		// фиксируем school_id, независимо от payload
		st.SchoolID = school.ID
	default:
		return false, errors.New("access denied")
	}

	rows, err := s.repo.Update(ctx, id, st)
	if err != nil {
		return false, err
	}
	if rows == 0 {
		return false, nil
	}

	// если поменялся класс — обновляем счётчики и для старого класса
	if existing.ClassID != st.ClassID {
		_ = s.UpdateCounts(ctx, existing.SchoolID, existing.ClassID)
	}
	_ = s.UpdateCounts(ctx, existing.SchoolID, st.ClassID)

	return true, nil
}

func (s *StudentService) Delete(ctx context.Context, id int, role string, userID int) (bool, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrStudentNotFound) {
			return false, nil
		}
		return false, err
	}

	switch role {
	case "roo":
		// может удалить любого
	case "school":
		school, err := s.schoolRepo.GetByUserID(ctx, userID)
		if err != nil {
			return false, err
		}
		if existing.SchoolID != school.ID {
			return false, errors.New("access denied")
		}
	default:
		return false, errors.New("access denied")
	}

	rows, err := s.repo.Delete(ctx, id)
	if err != nil {
		return false, err
	}
	if rows == 0 {
		return false, nil
	}

	_ = s.UpdateCounts(ctx, existing.SchoolID, existing.ClassID)
	return true, nil
}

func (s *StudentService) GetStats(ctx context.Context, role string) (map[string]int, error) {
	if role != "roo" {
		return nil, errors.New("access denied")
	}
	return s.repo.GetStats(ctx)
}

// ==== 🔧 Обновление счётчиков ====
func (s *StudentService) UpdateCounts(ctx context.Context, schoolID, classID int) error {
	count, err := s.repo.CountByClass(ctx, classID)
	if err != nil {
		return err
	}
	_, err = s.classRepo.DB().Exec(ctx, `UPDATE classes SET student_count=$1 WHERE id=$2`, count, classID)
	if err != nil {
		return err
	}

	_, err = s.schoolRepo.DB().Exec(ctx, `
		UPDATE schools
		SET student_count=(SELECT COUNT(*) FROM students WHERE school_id=$1)
		WHERE id=$1`, schoolID)
	return err
}
