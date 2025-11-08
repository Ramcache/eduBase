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

type StaffFilter struct {
	FullName        string
	Phone           string
	Position        string
	Education       string
	Category        string
	PedExperience   *int
	TotalExperience *int
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
		full_name, phone, position, education, category,
		ped_experience, total_experience, work_start, note, school_id
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	RETURNING id, created_at`
	return r.db.QueryRow(ctx, query,
		s.FullName, s.Phone, s.Position,
		s.Education, s.Category,
		s.PedExperience, s.TotalExperience,
		s.WorkStart, s.Note, s.SchoolID,
	).Scan(&s.ID, &s.CreatedAt)
}

func (r *StaffRepository) GetAll(ctx context.Context, schoolID *int, f StaffFilter) ([]models.Staff, error) {
	base := `
	SELECT id, full_name, phone, position, education, category,
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

	var list []models.Staff
	for rows.Next() {
		var s models.Staff
		if err := rows.Scan(
			&s.ID, &s.FullName, &s.Phone, &s.Position, &s.Education, &s.Category,
			&s.PedExperience, &s.TotalExperience, &s.WorkStart, &s.Note, &s.SchoolID, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *StaffRepository) Delete(ctx context.Context, id, schoolID int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM staff WHERE id=$1 AND school_id=$2`, id, schoolID)
	return err
}

func (r *StaffRepository) DB() *pgx.Conn {
	return r.db
}

func (r *StaffRepository) GetByID(ctx context.Context, id int) (*models.Staff, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, full_name, phone, position, education, category,
		       ped_experience, total_experience, work_start, note,
		       school_id, created_at
		FROM staff WHERE id=$1
	`, id)
	var s models.Staff
	if err := row.Scan(
		&s.ID, &s.FullName, &s.Phone, &s.Position, &s.Education, &s.Category,
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

func (r *StaffRepository) Update(ctx context.Context, id int, s *models.Staff, role string) (int64, error) {
	var res pgconn.CommandTag
	var err error
	if role == "roo" {
		res, err = r.db.Exec(ctx, `
			UPDATE staff
			SET full_name=$1, phone=$2, position=$3, education=$4,
			    category=$5, ped_experience=$6, total_experience=$7,
			    work_start=$8, note=$9
			WHERE id=$10`,
			s.FullName, s.Phone, s.Position, s.Education, s.Category,
			s.PedExperience, s.TotalExperience, s.WorkStart, s.Note, id,
		)
	} else {
		res, err = r.db.Exec(ctx, `
			UPDATE staff
			SET full_name=$1, phone=$2, position=$3, education=$4,
			    category=$5, ped_experience=$6, total_experience=$7,
			    work_start=$8, note=$9
			WHERE id=$10 AND school_id=$11`,
			s.FullName, s.Phone, s.Position, s.Education, s.Category,
			s.PedExperience, s.TotalExperience, s.WorkStart, s.Note, id, s.SchoolID,
		)
	}
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
