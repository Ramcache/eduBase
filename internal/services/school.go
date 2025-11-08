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

func (s *SchoolService) Update(ctx context.Context, id int, req *models.School) error {
	return s.repo.Update(ctx, id, req)
}

func (s *SchoolService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
