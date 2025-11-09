package repository

import (
	"context"
	"eduBase/internal/models"
	"github.com/jackc/pgx/v5"
)

type StatsRepository struct {
	db *pgx.Conn
}

func NewStatsRepository(db *pgx.Conn) *StatsRepository {
	return &StatsRepository{db: db}
}

func (r *StatsRepository) DB() *pgx.Conn { return r.db }

// GetSummary: если schoolID != nil — агрегаты по школе, иначе по всей системе
func (r *StatsRepository) GetSummary(ctx context.Context, schoolID *int) (*models.StatsSummary, error) {
	var q string
	var args []any

	if schoolID == nil {
		q = `
		WITH
		sch AS (SELECT COUNT(*)::int AS n FROM schools),
		c AS (SELECT COUNT(*)::int AS n FROM classes),
		stu AS (SELECT COUNT(*)::int AS n FROM students),
		t AS (SELECT COUNT(*)::int AS n FROM staff WHERE position ILIKE '%учител%'),
		st AS (SELECT COUNT(*)::int AS n FROM staff)
		SELECT sch.n, c.n, stu.n, t.n, st.n FROM sch,c,stu,t,st;
		`
	} else {
		q = `
		WITH
		sch AS (SELECT COUNT(*)::int AS n FROM schools WHERE id = $1),
		c AS (SELECT COUNT(*)::int AS n FROM classes  WHERE school_id = $1),
		stu AS (SELECT COUNT(*)::int AS n FROM students WHERE school_id = $1),
		t AS (SELECT COUNT(*)::int AS n FROM staff    WHERE school_id = $1 AND position ILIKE '%учител%'),
		st AS (SELECT COUNT(*)::int AS n FROM staff    WHERE school_id = $1)
		SELECT sch.n, c.n, stu.n, t.n, st.n FROM sch,c,stu,t,st;
		`
		args = append(args, *schoolID)
	}

	row := r.db.QueryRow(ctx, q, args...)
	var res models.StatsSummary
	if err := row.Scan(&res.Schools, &res.Classes, &res.Students, &res.Teachers, &res.StaffTotal); err != nil {
		return nil, err
	}
	return &res, nil
}

// Дополнительно: быстрая проверка существования школы (для валидации school_id у ROO)
func (r *StatsRepository) SchoolExists(ctx context.Context, id int) (bool, error) {
	var ok bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schools WHERE id=$1)`, id).Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}
