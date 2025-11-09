package services

import (
	"context"

	"eduBase/internal/models"
	"eduBase/internal/repository"
	"github.com/jackc/pgx/v5"
)

type StatsService struct {
	repo       *repository.StatsRepository
	schoolRepo *repository.SchoolRepository
}

func NewStatsService(repo *repository.StatsRepository, schoolRepo *repository.SchoolRepository) *StatsService {
	return &StatsService{repo: repo, schoolRepo: schoolRepo}
}

func (s *StatsService) RepoDB() *pgx.Conn { return s.repo.DB() }

func (s *StatsService) GetSummary(ctx context.Context, schoolID *int) (*models.StatsSummary, error) {
	return s.repo.GetSummary(ctx, schoolID)
}
