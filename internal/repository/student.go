package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"eduBase/internal/models"
	"github.com/jackc/pgx/v5"
)

type StudentRepository struct {
	db *pgx.Conn
}

func NewStudentRepository(db *pgx.Conn) *StudentRepository {
	return &StudentRepository{db: db}
}

var ErrStudentNotFound = errors.New("student not found")

func (r *StudentRepository) DB() *pgx.Conn {
	return r.db
}

type StudentFilter struct {
	FullName string
	Gender   string
	ClassID  *int

	GradeFrom *int
	GradeTo   *int

	Limit  int
	Offset int
}

// ===== CREATE =====
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

// ===== GET ALL (with class name, filters, pagination) =====
func (r *StudentRepository) GetAll(ctx context.Context, schoolID *int, f StudentFilter) ([]models.Student, error) {
	base := `
	SELECT s.id, s.full_name, s.birth_date, s.gender, s.phone, s.address, s.note,
	       s.class_id, c.name AS class_name, s.school_id, s.created_at
	FROM students s
	JOIN classes c ON c.id = s.class_id`
	var where []string
	var args []any
	i := 1

	if schoolID != nil {
		where = append(where, fmt.Sprintf("s.school_id=$%d", i))
		args = append(args, *schoolID)
		i++
	}
	if f.FullName != "" {
		where = append(where, fmt.Sprintf("LOWER(s.full_name) ILIKE $%d", i))
		args = append(args, "%"+strings.ToLower(f.FullName)+"%")
		i++
	}
	if f.Gender != "" {
		where = append(where, fmt.Sprintf("s.gender=$%d", i))
		args = append(args, f.Gender)
		i++
	}
	if f.ClassID != nil {
		where = append(where, fmt.Sprintf("s.class_id=$%d", i))
		args = append(args, *f.ClassID)
		i++
	}
	if f.GradeFrom != nil {
		where = append(where, fmt.Sprintf("c.grade >= $%d", i))
		args = append(args, *f.GradeFrom)
		i++
	}
	if f.GradeTo != nil {
		where = append(where, fmt.Sprintf("c.grade <= $%d", i))
		args = append(args, *f.GradeTo)
		i++
	}

	query := base
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY s.full_name"

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

	var list []models.Student
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(
			&s.ID, &s.FullName, &s.BirthDate, &s.Gender,
			&s.Phone, &s.Address, &s.Note,
			&s.ClassID, &s.ClassName, &s.SchoolID, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

// ===== GET BY ID (with class name) =====
func (r *StudentRepository) GetByID(ctx context.Context, id int) (*models.Student, error) {
	row := r.db.QueryRow(ctx, `
		SELECT s.id, s.full_name, s.birth_date, s.gender, s.phone, s.address, s.note,
		       s.class_id, c.name AS class_name, s.school_id, s.created_at
		FROM students s
		JOIN classes c ON c.id = s.class_id
		WHERE s.id=$1`, id)

	var s models.Student
	if err := row.Scan(
		&s.ID, &s.FullName, &s.BirthDate, &s.Gender,
		&s.Phone, &s.Address, &s.Note,
		&s.ClassID, &s.ClassName, &s.SchoolID, &s.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStudentNotFound
		}
		return nil, err
	}
	return &s, nil
}

// ===== UPDATE (без знания о роли / school_id) =====
func (r *StudentRepository) Update(ctx context.Context, id int, s *models.Student) (int64, error) {
	res, err := r.db.Exec(ctx, `
			UPDATE students
			SET full_name=$1, birth_date=$2, gender=$3, phone=$4,
			    address=$5, note=$6, class_id=$7
			WHERE id=$8`,
		s.FullName, s.BirthDate, s.Gender, s.Phone,
		s.Address, s.Note, s.ClassID, id,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

// ===== DELETE (без school_id) =====
func (r *StudentRepository) Delete(ctx context.Context, id int) (int64, error) {
	res, err := r.db.Exec(ctx, `DELETE FROM students WHERE id=$1`, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

// ===== COUNT BY CLASS =====
func (r *StudentRepository) CountByClass(ctx context.Context, classID int) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM students WHERE class_id=$1`, classID).Scan(&count)
	return count, err
}

// ===== STATS =====
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

func (r *StudentRepository) GetGradeStats(ctx context.Context, schoolID *int, gradeFrom, gradeTo *int) (int, int, error) {
	var (
		classCount    int
		studentCount  int
		args          []any
		whereClasses  []string
		whereStudents []string
		iClasses      = 1
		iStudents     = 1
	)

	// ==== классы ====
	classQuery := `SELECT COUNT(*) FROM classes c`
	if schoolID != nil {
		whereClasses = append(whereClasses, fmt.Sprintf("c.school_id=$%d", iClasses))
		args = append(args, *schoolID)
		iClasses++
	}
	if gradeFrom != nil {
		whereClasses = append(whereClasses, fmt.Sprintf("c.grade >= $%d", iClasses))
		args = append(args, *gradeFrom)
		iClasses++
	}
	if gradeTo != nil {
		whereClasses = append(whereClasses, fmt.Sprintf("c.grade <= $%d", iClasses))
		args = append(args, *gradeTo)
		iClasses++
	}
	if len(whereClasses) > 0 {
		classQuery += " WHERE " + strings.Join(whereClasses, " AND ")
	}

	// т.к. у нас общий args, проще отдельно выполнить запросы:
	// сделаем два отдельных блока с локальными args.

	// ---- классы ----
	var classArgs []any
	classArgs = append(classArgs, args...)
	if err := r.db.QueryRow(ctx, classQuery, classArgs...).Scan(&classCount); err != nil {
		return 0, 0, err
	}

	// ==== ученики ====
	studentQuery := `
		SELECT COUNT(*)
		FROM students s
		JOIN classes c ON c.id = s.class_id`
	args = nil
	if schoolID != nil {
		whereStudents = append(whereStudents, fmt.Sprintf("s.school_id=$%d", iStudents))
		args = append(args, *schoolID)
		iStudents++
	}
	if gradeFrom != nil {
		whereStudents = append(whereStudents, fmt.Sprintf("c.grade >= $%d", iStudents))
		args = append(args, *gradeFrom)
		iStudents++
	}
	if gradeTo != nil {
		whereStudents = append(whereStudents, fmt.Sprintf("c.grade <= $%d", iStudents))
		args = append(args, *gradeTo)
		iStudents++
	}
	if len(whereStudents) > 0 {
		studentQuery += " WHERE " + strings.Join(whereStudents, " AND ")
	}

	if err := r.db.QueryRow(ctx, studentQuery, args...).Scan(&studentCount); err != nil {
		return 0, 0, err
	}

	return classCount, studentCount, nil
}
