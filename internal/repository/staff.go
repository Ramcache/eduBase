package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"eduBase/internal/models"
	"github.com/jackc/pgx/v5"
)

type StaffFilter struct {
	FullName  string
	Phone     string
	Position  string
	Subject   string
	Education string
	Category  string

	// минимальные значения стажа (в годах)
	PedExperience   *int
	TotalExperience *int

	// пагинация
	Limit  int
	Offset int
}

type StaffRepository struct {
	db *pgx.Conn
}

func NewStaffRepository(db *pgx.Conn) *StaffRepository {
	return &StaffRepository{db: db}
}

var ErrStaffNotFound = errors.New("staff not found")

func (r *StaffRepository) Create(ctx context.Context, s *models.Staff) error {
	query := `
	INSERT INTO staff (
		full_name, phone, position, subject, education, category,
		ped_experience, total_experience, work_start, note, school_id
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	RETURNING id, created_at`
	return r.db.QueryRow(ctx, query,
		s.FullName, s.Phone, s.Position, s.Subject, s.Education, s.Category,
		s.PedExperience, s.TotalExperience, s.WorkStart, s.Note, s.SchoolID,
	).Scan(&s.ID, &s.CreatedAt)
}

// GetAll — ROO: schoolID == nil → все, School: только свои
func (r *StaffRepository) GetAll(ctx context.Context, schoolID *int, f StaffFilter) ([]models.Staff, error) {
	base := `
	SELECT id, full_name, phone, position, subject, education, category,
	       ped_experience, total_experience, work_start, note, school_id, created_at
	FROM staff`
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
	if f.Phone != "" {
		where = append(where, fmt.Sprintf("phone ILIKE $%d", i))
		args = append(args, "%"+f.Phone+"%")
		i++
	}
	if f.Position != "" {
		where = append(where, fmt.Sprintf("LOWER(position) ILIKE $%d", i))
		args = append(args, "%"+strings.ToLower(f.Position)+"%")
		i++
	}
	if f.Subject != "" {
		where = append(where, fmt.Sprintf("LOWER(subject) ILIKE $%d", i))
		args = append(args, "%"+strings.ToLower(f.Subject)+"%")
		i++
	}
	if f.Education != "" {
		where = append(where, fmt.Sprintf("LOWER(education) ILIKE $%d", i))
		args = append(args, "%"+strings.ToLower(f.Education)+"%")
		i++
	}
	if f.Category != "" {
		where = append(where, fmt.Sprintf("LOWER(category) ILIKE $%d", i))
		args = append(args, "%"+strings.ToLower(f.Category)+"%")
		i++
	}
	if f.PedExperience != nil {
		where = append(where, fmt.Sprintf("ped_experience >= $%d", i))
		args = append(args, *f.PedExperience)
		i++
	}
	if f.TotalExperience != nil {
		where = append(where, fmt.Sprintf("total_experience >= $%d", i))
		args = append(args, *f.TotalExperience)
		i++
	}

	query := base
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY full_name"

	// пагинация
	if f.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", i)
		args = append(args, f.Limit)
		i++
	}
	if f.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", i)
		args = append(args, f.Offset)
		i++
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Staff
	for rows.Next() {
		var s models.Staff
		if err := rows.Scan(
			&s.ID, &s.FullName, &s.Phone, &s.Position, &s.Subject, &s.Education,
			&s.Category, &s.PedExperience, &s.TotalExperience,
			&s.WorkStart, &s.Note, &s.SchoolID, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

// Delete — roo: schoolID == nil → без условия по school_id, school: только свои
func (r *StaffRepository) Delete(ctx context.Context, id int, schoolID *int) (int64, error) {
	var (
		query string
		args  []any
	)

	if schoolID != nil {
		query = `DELETE FROM staff WHERE id=$1 AND school_id=$2`
		args = append(args, id, *schoolID)
	} else {
		query = `DELETE FROM staff WHERE id=$1`
		args = append(args, id)
	}

	res, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func (r *StaffRepository) DB() *pgx.Conn {
	return r.db
}

func (r *StaffRepository) GetByID(ctx context.Context, id int) (*models.Staff, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, full_name, phone, position, subject, education, category,
		       ped_experience, total_experience, work_start, note,
		       school_id, created_at
		FROM staff WHERE id=$1
	`, id)
	var s models.Staff
	if err := row.Scan(
		&s.ID, &s.FullName, &s.Phone, &s.Position, &s.Subject, &s.Education, &s.Category,
		&s.PedExperience, &s.TotalExperience, &s.WorkStart, &s.Note,
		&s.SchoolID, &s.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStaffNotFound
		}
		return nil, err
	}
	return &s, nil
}

// Update — вообще не трогаем school_id, он контролируется выше (в сервисе)
func (r *StaffRepository) Update(ctx context.Context, id int, s *models.Staff) (int64, error) {
	res, err := r.db.Exec(ctx, `
		UPDATE staff
		SET full_name=$1,
		    phone=$2,
		    position=$3,
		    subject=$4,
		    education=$5,
		    category=$6,
		    ped_experience=$7,
		    total_experience=$8,
		    work_start=$9,
		    note=$10
		WHERE id=$11`,
		s.FullName, s.Phone, s.Position, s.Subject,
		s.Education, s.Category, s.PedExperience,
		s.TotalExperience, s.WorkStart, s.Note, id,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

// агрегированная статистика (для ROO)
func (r *StaffRepository) GetStats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)
	rows, err := r.db.Query(ctx, `
		SELECT position, COUNT(*) FROM staff GROUP BY position
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var pos string
		var count int
		if err := rows.Scan(&pos, &count); err != nil {
			return nil, err
		}
		stats[pos] = count
	}
	return stats, nil
}
