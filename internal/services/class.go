package services

import (
	"context"
	"github.com/jackc/pgx/v5"

	"eduBase/internal/models"
	"eduBase/internal/repository"
)

type ClassService struct {
	repo *repository.ClassRepository
	db   *pgx.Conn
}

func NewClassService(repo *repository.ClassRepository) *ClassService {
	return &ClassService{repo: repo, db: repo.DB()}
}

func (s *ClassService) RepoDB() *pgx.Conn {
	return s.db
}

func (s *ClassService) Create(ctx context.Context, c *models.Class) error {
	return s.repo.Create(ctx, c)
}

func (s *ClassService) GetAll(ctx context.Context) ([]models.Class, error) {
	return s.repo.GetAll(ctx)
}

func (s *ClassService) GetBySchool(ctx context.Context, schoolID int) ([]models.Class, error) {
	return s.repo.GetBySchool(ctx, schoolID)
}

func (s *ClassService) Update(ctx context.Context, id int, c *models.Class, role string) (bool, error) {
	rows, err := s.repo.Update(ctx, id, c, role)
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (s *ClassService) Delete(ctx context.Context, id, schoolID int) error {
	return s.repo.Delete(ctx, id, schoolID)
}

func (s *ClassService) GetByID(ctx context.Context, id int) (*models.Class, error) {
	return s.repo.GetByID(ctx, id)
}
