package services

import (
	"context"

	"eduBase/internal/models"
	"eduBase/internal/repository"
)

type SchoolService struct {
	repo *repository.SchoolRepository
}

func NewSchoolService(repo *repository.SchoolRepository) *SchoolService {
	return &SchoolService{repo: repo}
}

func (s *SchoolService) GetAll(ctx context.Context) ([]models.School, error) {
	return s.repo.GetAll(ctx)
}

func (s *SchoolService) GetByID(ctx context.Context, id int) (*models.School, error) {
	return s.repo.GetByID(ctx, id)
}

// Update — используется ROO. Меняем только name/director, остальные поля руками не трогаем.
func (s *SchoolService) Update(ctx context.Context, id int, req *models.School) error {
	// Берём текущую школу из БД, чтобы не потерять счётчики и прочие поля.
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Разрешаем менять только эти поля:
	existing.Name = req.Name
	existing.Director = req.Director

	return s.repo.Update(ctx, id, existing)
}

func (s *SchoolService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *SchoolService) GetByUserID(ctx context.Context, userID int) (*models.School, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// UpdateByUserID — школа обновляет только себя и тоже только name/director.
func (s *SchoolService) UpdateByUserID(ctx context.Context, userID int, req *models.School) error {
	school, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Разрешаем менять только эти поля:
	school.Name = req.Name
	school.Director = req.Director

	return s.repo.Update(ctx, school.ID, school)
}
