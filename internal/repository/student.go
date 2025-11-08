package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"strings"

	"eduBase/internal/models"
	"github.com/jackc/pgx/v5"
)

type StudentFilter struct {
	FullName string
	Gender   string
	ClassID  *int
}

type StudentRepository struct {
	db *pgx.Conn
}

func NewStudentRepository(db *pgx.Conn) *StudentRepository {
	return &StudentRepository{db: db}
}

var ErrStudentNotFound = errors.New("student not found")

func (r *StudentRepository) Create(ctx context.Context, s *models.Student) error {
	query := `
		INSERT INTO students (full_name, birth_date, gender, phone, address, note, class_id, school_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, created_at`
	return r.db.QueryRow(ctx, query,
		s.FullName, s.BirthDate, s.Gender, s.Phone, s.Address, s.Note,
		s.ClassID, s.SchoolID,
	).Scan(&s.ID, &s.CreatedAt)
}

func (r *StudentRepository) GetAll(ctx context.Context, schoolID *int, f StudentFilter) ([]models.Student, error) {
	base := `
	SELECT id, full_name, birth_date, gender, phone, address, note, class_id, school_id, created_at
	FROM students`
	var where []string
	var args []any
	i := 1

	if schoolID != nil {
		where = append(where, fmt.Sprintf("school_id=$%d", i))
		args = append(args, *schoolID)
		i++
	}
	if f.FullName != "" {
		where = append(where, fmt.Sprintf("LOWER(full_name) ILIKE $%d", i))
		args = append(args, "%"+strings.ToLower(f.FullName)+"%")
		i++
	}
	if f.Gender != "" {
		where = append(where, fmt.Sprintf("gender=$%d", i))
		args = append(args, f.Gender)
		i++
	}
	if f.ClassID != nil {
		where = append(where, fmt.Sprintf("class_id=$%d", i))
		args = append(args, *f.ClassID)
		i++
	}

	query := base
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY full_name"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Student
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(&s.ID, &s.FullName, &s.BirthDate, &s.Gender, &s.Phone, &s.Address, &s.Note, &s.ClassID, &s.SchoolID, &s.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *StudentRepository) Delete(ctx context.Context, id int, schoolID int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM students WHERE id=$1 AND school_id=$2`, id, schoolID)
	return err
}

func (r *StudentRepository) CountByClass(ctx context.Context, classID int) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM students WHERE class_id=$1`, classID).Scan(&count)
	return count, err
}

func (r *StudentRepository) GetByID(ctx context.Context, id int) (*models.Student, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, full_name, birth_date, gender, phone, address, note, class_id, school_id, created_at
		FROM students WHERE id=$1`, id)

	var s models.Student
	if err := row.Scan(
		&s.ID, &s.FullName, &s.BirthDate, &s.Gender, &s.Phone,
		&s.Address, &s.Note, &s.ClassID, &s.SchoolID, &s.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStudentNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *StudentRepository) Update(ctx context.Context, id int, s *models.Student, role string) (int64, error) {
	var res pgconn.CommandTag
	var err error
	if role == "roo" {
		res, err = r.db.Exec(ctx, `
			UPDATE students
			SET full_name=$1, birth_date=$2, gender=$3, phone=$4, address=$5, note=$6, class_id=$7
			WHERE id=$8`,
			s.FullName, s.BirthDate, s.Gender, s.Phone, s.Address, s.Note, s.ClassID, id)
	} else {
		res, err = r.db.Exec(ctx, `
			UPDATE students
			SET full_name=$1, birth_date=$2, gender=$3, phone=$4, address=$5, note=$6, class_id=$7
			WHERE id=$8 AND school_id=$9`,
			s.FullName, s.BirthDate, s.Gender, s.Phone, s.Address, s.Note, s.ClassID, id, s.SchoolID)
	}
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func (r *StudentRepository) GetStats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)
	rows, err := r.db.Query(ctx, `SELECT gender, COUNT(*) FROM students GROUP BY gender`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var g string
		var c int
		if err := rows.Scan(&g, &c); err != nil {
			return nil, err
		}
		stats[g] = c
	}
	return stats, nil
}
