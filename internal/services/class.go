package services

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"

	"eduBase/internal/models"
	"eduBase/internal/repository"
)

type ClassService struct {
	repo       *repository.ClassRepository
	schoolRepo *repository.SchoolRepository
}

func NewClassService(repo *repository.ClassRepository, sr *repository.SchoolRepository) *ClassService {
	return &ClassService{repo: repo, schoolRepo: sr}
}

// RepoDB оставил, если вдруг где-то ещё используется
func (s *ClassService) RepoDB() *pgx.Conn {
	return s.repo.DB()
}

// Create — только school; school_id берётся из user_id
func (s *ClassService) Create(ctx context.Context, c *models.Class, role string, userID int) error {
	if role != "school" {
		return errors.New("access denied")
	}

	school, err := s.schoolRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	c.SchoolID = school.ID

	return s.repo.Create(ctx, c)
}

// GetAll — roo → все классы, school → только свои
func (s *ClassService) GetAll(ctx context.Context, role string, userID int) ([]models.Class, error) {
	var schoolID *int

	switch role {
	case "roo":
		// видит все
	case "school":
		school, err := s.schoolRepo.GetByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		schoolID = &school.ID
	default:
		return nil, errors.New("access denied")
	}

	return s.repo.GetAll(ctx, schoolID)
}

// GetBySchool оставил, если надо где-то ещё
func (s *ClassService) GetBySchool(ctx context.Context, schoolID int) ([]models.Class, error) {
	return s.repo.GetAll(ctx, &schoolID)
}

func (s *ClassService) GetByID(ctx context.Context, id int, role string, userID int) (*models.Class, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	switch role {
	case "roo":
		// всё ок
	case "school":
		school, err := s.schoolRepo.GetByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if c.SchoolID != school.ID {
			return nil, errors.New("access denied")
		}
	default:
		return nil, errors.New("access denied")
	}

	return c, nil
}

func (s *ClassService) Update(ctx context.Context, id int, c *models.Class, role string, userID int) (bool, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrClassNotFound) {
			return false, nil
		}
		return false, err
	}

	switch role {
	case "roo":
		// может обновлять любой класс, но school_id не трогаем
	case "school":
		school, err := s.schoolRepo.GetByUserID(ctx, userID)
		if err != nil {
			return false, err
		}
		if existing.SchoolID != school.ID {
			return false, errors.New("access denied")
		}
		// school_id фиксируем, даже если в payload кто-то подменил
		c.SchoolID = school.ID
	default:
		return false, errors.New("access denied")
	}

	rows, err := s.repo.Update(ctx, id, c)
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (s *ClassService) Delete(ctx context.Context, id int, role string, userID int) (bool, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrClassNotFound) {
			return false, nil
		}
		return false, err
	}

	switch role {
	case "roo":
		// может удалить любой
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
	return rows > 0, nil
}
