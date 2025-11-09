package repository

import (
	"context"
	"errors"

	"eduBase/internal/models"
	"github.com/jackc/pgx/v5"
)

var ErrSchoolNotFound = errors.New("school not found")

type SchoolRepository struct {
	db *pgx.Conn
}

func NewSchoolRepository(db *pgx.Conn) *SchoolRepository {
	return &SchoolRepository{db: db}
}

func (r *SchoolRepository) Create(ctx context.Context, s *models.School, userID int) error {
	query := `
		INSERT INTO schools (name, director, user_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query, s.Name, s.Director, userID).
		Scan(&s.ID, &s.CreatedAt)
}

func (r *SchoolRepository) GetAll(ctx context.Context) ([]models.School, error) {
	rows, err := r.db.Query(ctx, `
		SELECT 
			s.id, s.name, s.director, s.class_count, s.student_count, s.user_id, s.created_at,
			u.id, u.email, u.password, u.role
		FROM schools s
		JOIN users u ON u.id = s.user_id
		ORDER BY s.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.School
	for rows.Next() {
		var s models.School
		var u models.UserInfo

		if err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.Director,
			&s.ClassCount,
			&s.StudentCount,
			&s.UserID,
			&s.CreatedAt,
			&u.ID,
			&u.Email,
			&u.Password,
			&u.Role,
		); err != nil {
			return nil, err
		}

		s.User = &u
		list = append(list, s)
	}
	return list, nil
}

func (r *SchoolRepository) GetByID(ctx context.Context, id int) (*models.School, error) {
	row := r.db.QueryRow(ctx, `
		SELECT 
			s.id, s.name, s.director, s.class_count, s.student_count, s.user_id, s.created_at,
			u.id, u.email, u.password, u.role
		FROM schools s
		JOIN users u ON u.id = s.user_id
		WHERE s.id = $1
	`, id)

	var s models.School
	var u models.UserInfo

	if err := row.Scan(
		&s.ID,
		&s.Name,
		&s.Director,
		&s.ClassCount,
		&s.StudentCount,
		&s.UserID,
		&s.CreatedAt,
		&u.ID,
		&u.Email,
		&u.Password,
		&u.Role,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSchoolNotFound
		}
		return nil, err
	}

	s.User = &u
	return &s, nil
}

func (r *SchoolRepository) Update(ctx context.Context, id int, s *models.School) error {
	_, err := r.db.Exec(ctx, `
		UPDATE schools
		SET name=$1, director=$2
		WHERE id=$3
	`, s.Name, s.Director, id)
	return err
}

func (r *SchoolRepository) Delete(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM schools WHERE id=$1`, id)
	return err
}

func (r *SchoolRepository) GetByUserID(ctx context.Context, userID int) (*models.School, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, director, class_count, student_count, created_at
		FROM schools WHERE user_id=$1
	`, userID)

	var s models.School
	if err := row.Scan(&s.ID, &s.Name, &s.Director, &s.ClassCount, &s.StudentCount, &s.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSchoolNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *SchoolRepository) DB() *pgx.Conn {
	return r.db
}
