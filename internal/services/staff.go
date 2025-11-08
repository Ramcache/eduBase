package services

import (
	"context"

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

func (s *StaffService) Create(ctx context.Context, staff *models.Staff) error {
	return s.repo.Create(ctx, staff)
}

func (s *StaffService) GetAll(ctx context.Context, schoolID *int, f repository.StaffFilter) ([]models.Staff, error) {
	return s.repo.GetAll(ctx, schoolID, f)
}

func (s *StaffService) Delete(ctx context.Context, id, schoolID int) error {
	return s.repo.Delete(ctx, id, schoolID)
}

func (s *StaffService) GetByID(ctx context.Context, id int) (*models.Staff, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *StaffService) Update(ctx context.Context, id int, staff *models.Staff, role string) (bool, error) {
	rows, err := s.repo.Update(ctx, id, staff, role)
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (s *StaffService) GetStats(ctx context.Context) (map[string]int, error) {
	return s.repo.GetStats(ctx)
}
