package repository

import (
	"context"
	"errors"

	"eduBase/internal/models"
	"github.com/jackc/pgx/v5"
)

var ErrClassNotFound = errors.New("class not found")

type ClassRepository struct {
	db *pgx.Conn
}

func NewClassRepository(db *pgx.Conn) *ClassRepository {
	return &ClassRepository{db: db}
}

func (r *ClassRepository) DB() *pgx.Conn { return r.db }

func (r *ClassRepository) Create(ctx context.Context, c *models.Class) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO classes (name, grade, school_id)
		 VALUES ($1,$2,$3) RETURNING id,created_at`,
		c.Name, c.Grade, c.SchoolID,
	).Scan(&c.ID, &c.CreatedAt)
}

// GetAll — если schoolID == nil → все классы, иначе только этой школы
func (r *ClassRepository) GetAll(ctx context.Context, schoolID *int) ([]models.Class, error) {
	var rows pgx.Rows
	var err error

	if schoolID != nil {
		rows, err = r.db.Query(ctx,
			`SELECT id,name,grade,school_id,student_count,created_at
			 FROM classes WHERE school_id=$1 ORDER BY id`, *schoolID)
	} else {
		rows, err = r.db.Query(ctx,
			`SELECT id,name,grade,school_id,student_count,created_at
			 FROM classes ORDER BY id`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []models.Class
	for rows.Next() {
		var c models.Class
		if err := rows.Scan(&c.ID, &c.Name, &c.Grade, &c.SchoolID, &c.StudentCount, &c.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, c)
	}
	return res, nil
}

// для совместимости, если где-то ещё используется
func (r *ClassRepository) GetBySchool(ctx context.Context, schoolID int) ([]models.Class, error) {
	return r.GetAll(ctx, &schoolID)
}

func (r *ClassRepository) Update(ctx context.Context, id int, c *models.Class) (int64, error) {
	res, err := r.db.Exec(ctx,
		`UPDATE classes SET name=$1, grade=$2 WHERE id=$3`,
		c.Name, c.Grade, id,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func (r *ClassRepository) Delete(ctx context.Context, id int) (int64, error) {
	res, err := r.db.Exec(ctx, `DELETE FROM classes WHERE id=$1`, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func (r *ClassRepository) GetByID(ctx context.Context, id int) (*models.Class, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, name, grade, school_id, student_count, created_at
		FROM classes WHERE id=$1
	`, id)
	var c models.Class
	if err := row.Scan(&c.ID, &c.Name, &c.Grade, &c.SchoolID, &c.StudentCount, &c.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrClassNotFound
		}
		return nil, err
	}
	return &c, nil
}
