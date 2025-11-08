package services

import (
	"context"

	"eduBase/internal/models"
	"eduBase/internal/repository"
	"github.com/jackc/pgx/v5"
)

type StudentService struct {
	repo       *repository.StudentRepository
	classRepo  *repository.ClassRepository
	schoolRepo *repository.SchoolRepository
}

func NewStudentService(r *repository.StudentRepository, cr *repository.ClassRepository, sr *repository.SchoolRepository) *StudentService {
	return &StudentService{repo: r, classRepo: cr, schoolRepo: sr}
}

// ==== 🔧 Геттеры для БД ====
func (s *StudentService) SchoolRepoDB() *pgx.Conn {
	return s.schoolRepo.DB()
}

func (s *StudentService) ClassRepoDB() *pgx.Conn {
	return s.classRepo.DB()
}

// ==== 🔧 CRUD ====
func (s *StudentService) Create(ctx context.Context, st *models.Student) error {
	if err := s.repo.Create(ctx, st); err != nil {
		return err
	}
	return s.UpdateCounts(ctx, st.SchoolID, st.ClassID)
}

func (s *StudentService) GetAll(ctx context.Context, schoolID *int, f repository.StudentFilter) ([]models.Student, error) {
	return s.repo.GetAll(ctx, schoolID, f)
}

func (s *StudentService) Delete(ctx context.Context, id, schoolID, classID int) error {
	if err := s.repo.Delete(ctx, id, schoolID); err != nil {
		return err
	}
	return s.UpdateCounts(ctx, schoolID, classID)
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

func (s *StudentService) GetByID(ctx context.Context, id int) (*models.Student, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *StudentService) Update(ctx context.Context, id int, st *models.Student, role string) (bool, error) {
	rows, err := s.repo.Update(ctx, id, st, role)
	if err != nil {
		return false, err
	}
	if rows > 0 {
		_ = s.UpdateCounts(ctx, st.SchoolID, st.ClassID)
	}
	return rows > 0, nil
}

func (s *StudentService) GetStats(ctx context.Context) (map[string]int, error) {
	return s.repo.GetStats(ctx)
}
