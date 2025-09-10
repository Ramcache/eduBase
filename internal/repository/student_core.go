package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"eduBase/internal/logger"
	"eduBase/internal/models"
)

type StudentFilters struct {
	Q                 string
	SchoolID          *int
	ClassLabel        *string
	Status            *string
	AdmissionYearFrom *int
	AdmissionYearTo   *int
	BirthDateFrom     *time.Time
	BirthDateTo       *time.Time
}

type StudentCoreRepository interface {
	Create(ctx context.Context, s *models.StudentCore) (int, error)
	Get(ctx context.Context, id int) (*models.StudentCore, error)
	Update(ctx context.Context, s *models.StudentCore) error
	SoftDelete(ctx context.Context, id int, by int) error
	List(ctx context.Context, f StudentFilters, limit, offset int) ([]models.StudentListItem, int, error)
	ListFull(ctx context.Context, f StudentFilters, limit, offset int) ([]models.StudentCore, int, error)
	FindIDByStudentNumber(ctx context.Context, studentNumber string) (int, bool, error)
}

type studentCoreRepo struct {
	db *pgxpool.Pool
}

func NewStudentCoreRepo(db *pgxpool.Pool) StudentCoreRepository {
	return &studentCoreRepo{db: db}
}

func (r *studentCoreRepo) Create(ctx context.Context, s *models.StudentCore) (int, error) {
	var id int
	err := r.db.QueryRow(ctx, `
INSERT INTO students_core (
  student_number,last_name,first_name,middle_name,birth_date,gender,citizenship,
  school_id,class_label,admission_year,status,
  reg_address,fact_address,student_phone,student_email,
  created_by
) VALUES (
  $1,$2,$3,$4,$5,$6,$7,
  $8,$9,$10,$11,
  $12,$13,$14,$15,
  $16
) RETURNING id`,
		s.StudentNumber, s.LastName, s.FirstName, s.MiddleName, s.BirthDate, s.Gender, s.Citizenship,
		s.SchoolID, s.ClassLabel, s.AdmissionYear, s.Status,
		s.RegAddress, s.FactAddress, s.StudentPhone, s.StudentEmail,
		s.CreatedBy,
	).Scan(&id)
	if err != nil {
		logger.Log.Error("student_core create", logger.Err(err))
		return 0, err
	}
	return id, nil
}

func (r *studentCoreRepo) Get(ctx context.Context, id int) (*models.StudentCore, error) {
	row := r.db.QueryRow(ctx, `
	SELECT id,student_number,last_name,first_name,middle_name,birth_date,gender,citizenship,
		   school_id,class_label,admission_year,status,
		   reg_address,fact_address,student_phone,student_email,
		   created_at,updated_at,deleted_at,created_by,updated_by
	FROM students_core
	WHERE id=$1 AND deleted_at IS NULL`, id)

	var m models.StudentCore
	err := row.Scan(
		&m.ID, &m.StudentNumber, &m.LastName, &m.FirstName, &m.MiddleName, &m.BirthDate, &m.Gender, &m.Citizenship,
		&m.SchoolID, &m.ClassLabel, &m.AdmissionYear, &m.Status,
		&m.RegAddress, &m.FactAddress, &m.StudentPhone, &m.StudentEmail,
		&m.CreatedAt, &m.UpdatedAt, &m.DeletedAt, &m.CreatedBy, &m.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *studentCoreRepo) Update(ctx context.Context, s *models.StudentCore) error {
	_, err := r.db.Exec(ctx, `
UPDATE students_core SET
  student_number=$1,last_name=$2,first_name=$3,middle_name=$4,
  birth_date=$5,gender=$6,citizenship=$7,
  school_id=$8,class_label=$9,admission_year=$10,status=$11,
  reg_address=$12,fact_address=$13,student_phone=$14,student_email=$15,
  updated_at=now(), updated_by=$16
WHERE id=$17 AND deleted_at IS NULL`,
		s.StudentNumber, s.LastName, s.FirstName, s.MiddleName,
		s.BirthDate, s.Gender, s.Citizenship,
		s.SchoolID, s.ClassLabel, s.AdmissionYear, s.Status,
		s.RegAddress, s.FactAddress, s.StudentPhone, s.StudentEmail,
		*s.UpdatedBy, s.ID,
	)
	if err != nil {
		logger.Log.Error("student_core update", logger.Err(err))
	}
	return err
}

func (r *studentCoreRepo) SoftDelete(ctx context.Context, id int, by int) error {
	_, err := r.db.Exec(ctx, `UPDATE students_core SET deleted_at=now(), updated_by=$1 WHERE id=$2 AND deleted_at IS NULL`, by, id)
	return err
}

func (r *studentCoreRepo) List(ctx context.Context, f StudentFilters, limit, offset int) ([]models.StudentListItem, int, error) {
	where := "WHERE deleted_at IS NULL"
	var args []any
	i := 1

	if f.Q != "" {
		where += fmt.Sprintf(" AND (last_name || ' ' || first_name || ' ' || COALESCE(middle_name,'')) ILIKE $%d", i)
		args = append(args, "%"+f.Q+"%")
		i++
	}
	if f.SchoolID != nil {
		where += fmt.Sprintf(" AND school_id = $%d", i)
		args = append(args, *f.SchoolID)
		i++
	}
	if f.ClassLabel != nil && *f.ClassLabel != "" {
		where += fmt.Sprintf(" AND class_label = $%d", i)
		args = append(args, *f.ClassLabel)
		i++
	}
	if f.Status != nil && *f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, *f.Status)
		i++
	}
	if f.AdmissionYearFrom != nil {
		where += fmt.Sprintf(" AND admission_year >= $%d", i)
		args = append(args, *f.AdmissionYearFrom)
		i++
	}
	if f.AdmissionYearTo != nil {
		where += fmt.Sprintf(" AND admission_year <= $%d", i)
		args = append(args, *f.AdmissionYearTo)
		i++
	}
	if f.BirthDateFrom != nil {
		where += fmt.Sprintf(" AND birth_date >= $%d", i)
		args = append(args, *f.BirthDateFrom)
		i++
	}
	if f.BirthDateTo != nil {
		where += fmt.Sprintf(" AND birth_date <= $%d", i)
		args = append(args, *f.BirthDateTo)
		i++
	}

	var total int
	countSQL := "SELECT count(*) FROM students_core " + where
	if err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	listSQL := strings.Builder{}
	listSQL.WriteString(`
SELECT id,
  student_number,
  (last_name || ' ' || first_name || ' ' || COALESCE(middle_name,'')) AS full_name,
  birth_date, gender, school_id, class_label, admission_year, status, created_at
FROM students_core `)
	listSQL.WriteString(where)
	listSQL.WriteString(fmt.Sprintf(" ORDER BY last_name, first_name LIMIT $%d OFFSET $%d", i, i+1))

	rows, err := r.db.Query(ctx, listSQL.String(), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []models.StudentListItem
	for rows.Next() {
		var it models.StudentListItem
		if err := rows.Scan(
			&it.ID, &it.StudentNumber, &it.FullName, &it.BirthDate, &it.Gender,
			&it.SchoolID, &it.ClassLabel, &it.AdmissionYear, &it.Status, &it.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

func (r *studentCoreRepo) ListFull(ctx context.Context, f StudentFilters, limit, offset int) ([]models.StudentCore, int, error) {
	where := "WHERE deleted_at IS NULL"
	var args []any
	i := 1

	if f.Q != "" {
		where += fmt.Sprintf(" AND (last_name || ' ' || first_name || ' ' || COALESCE(middle_name,'')) ILIKE $%d", i)
		args = append(args, "%"+f.Q+"%")
		i++
	}
	if f.SchoolID != nil {
		where += fmt.Sprintf(" AND school_id = $%d", i)
		args = append(args, *f.SchoolID)
		i++
	}
	if f.ClassLabel != nil && *f.ClassLabel != "" {
		where += fmt.Sprintf(" AND class_label = $%d", i)
		args = append(args, *f.ClassLabel)
		i++
	}
	if f.Status != nil && *f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, *f.Status)
		i++
	}
	if f.AdmissionYearFrom != nil {
		where += fmt.Sprintf(" AND admission_year >= $%d", i)
		args = append(args, *f.AdmissionYearFrom)
		i++
	}
	if f.AdmissionYearTo != nil {
		where += fmt.Sprintf(" AND admission_year <= $%d", i)
		args = append(args, *f.AdmissionYearTo)
		i++
	}
	if f.BirthDateFrom != nil {
		where += fmt.Sprintf(" AND birth_date >= $%d", i)
		args = append(args, *f.BirthDateFrom)
		i++
	}
	if f.BirthDateTo != nil {
		where += fmt.Sprintf(" AND birth_date <= $%d", i)
		args = append(args, *f.BirthDateTo)
		i++
	}

	var total int
	if err := r.db.QueryRow(ctx, "SELECT count(*) FROM students_core "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	sql := fmt.Sprintf(`
SELECT id,student_number,last_name,first_name,middle_name,birth_date,gender,citizenship,
       school_id,class_label,admission_year,status,
       reg_address,fact_address,student_phone,student_email,
       created_at,updated_at,deleted_at,created_by,updated_by
FROM students_core
%s
ORDER BY last_name, first_name
LIMIT $%d OFFSET $%d`, where, i, i+1)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []models.StudentCore
	for rows.Next() {
		var m models.StudentCore
		if err := rows.Scan(
			&m.ID, &m.StudentNumber, &m.LastName, &m.FirstName, &m.MiddleName, &m.BirthDate, &m.Gender, &m.Citizenship,
			&m.SchoolID, &m.ClassLabel, &m.AdmissionYear, &m.Status,
			&m.RegAddress, &m.FactAddress, &m.StudentPhone, &m.StudentEmail,
			&m.CreatedAt, &m.UpdatedAt, &m.DeletedAt, &m.CreatedBy, &m.UpdatedBy,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, m)
	}
	return out, total, rows.Err()
}

func (r *studentCoreRepo) FindIDByStudentNumber(ctx context.Context, studentNumber string) (int, bool, error) {
	var id int
	err := r.db.QueryRow(ctx, `SELECT id FROM students_core WHERE student_number=$1 AND deleted_at IS NULL`, studentNumber).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return id, true, nil
}
