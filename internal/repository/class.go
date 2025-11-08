package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"

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

func (r *ClassRepository) Create(ctx context.Context, c *models.Class) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO classes (name, grade, school_id)
		 VALUES ($1,$2,$3) RETURNING id,created_at`,
		c.Name, c.Grade, c.SchoolID,
	).Scan(&c.ID, &c.CreatedAt)
}

func (r *ClassRepository) GetAll(ctx context.Context) ([]models.Class, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id,name,grade,school_id,student_count,created_at
		 FROM classes ORDER BY id`)
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

func (r *ClassRepository) GetBySchool(ctx context.Context, schoolID int) ([]models.Class, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id,name,grade,school_id,student_count,created_at
		 FROM classes WHERE school_id=$1 ORDER BY id`, schoolID)
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

func (r *ClassRepository) Update(ctx context.Context, id int, c *models.Class, role string) (int64, error) {
	var res pgconn.CommandTag
	var err error

	if role == "roo" {
		res, err = r.db.Exec(ctx,
			`UPDATE classes SET name=$1, grade=$2 WHERE id=$3`,
			c.Name, c.Grade, id,
		)
	} else {
		res, err = r.db.Exec(ctx,
			`UPDATE classes SET name=$1, grade=$2 WHERE id=$3 AND school_id=$4`,
			c.Name, c.Grade, id, c.SchoolID,
		)
	}

	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func (r *ClassRepository) Delete(ctx context.Context, id int, schoolID int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM classes WHERE id=$1 AND school_id=$2`, id, schoolID)
	return err
}

func (r *ClassRepository) DB() *pgx.Conn { return r.db }
