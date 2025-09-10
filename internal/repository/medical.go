package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eduBase/internal/models"
)

type MedicalRepository interface {
	Upsert(ctx context.Context, studentID int, benefits, notes *string, health *int, allergies, activities *string) error
	Get(ctx context.Context, studentID int) (*models.StudentMedical, error)
	BulkByStudentIDs(ctx context.Context, ids []int) (map[int]*models.StudentMedical, error)
}

type medicalRepo struct{ db *pgxpool.Pool }

func NewMedicalRepo(db *pgxpool.Pool) MedicalRepository { return &medicalRepo{db: db} }

func (r *medicalRepo) Upsert(ctx context.Context, studentID int, benefits, notes *string, health *int, allergies, activities *string) error {
	_, err := r.db.Exec(ctx, `
INSERT INTO students_medical(student_id,benefits,medical_notes,health_group,allergies,activities)
VALUES($1,$2,$3,$4,$5,$6)
ON CONFLICT (student_id) DO UPDATE SET
 benefits=EXCLUDED.benefits,
 medical_notes=EXCLUDED.medical_notes,
 health_group=EXCLUDED.health_group,
 allergies=EXCLUDED.allergies,
 activities=EXCLUDED.activities,
 updated_at=now()`,
		studentID, benefits, notes, health, allergies, activities)
	return err
}

func (r *medicalRepo) Get(ctx context.Context, studentID int) (*models.StudentMedical, error) {
	row := r.db.QueryRow(ctx, `
SELECT student_id, benefits, medical_notes, health_group, allergies, activities, updated_at
FROM students_medical WHERE student_id=$1`, studentID)

	var m models.StudentMedical
	if err := row.Scan(&m.StudentID, &m.Benefits, &m.MedicalNotes, &m.HealthGroup, &m.Allergies, &m.Activities, &m.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *medicalRepo) BulkByStudentIDs(ctx context.Context, ids []int) (map[int]*models.StudentMedical, error) {
	if len(ids) == 0 {
		return map[int]*models.StudentMedical{}, nil
	}
	ints := make([]int32, 0, len(ids))
	for _, v := range ids {
		ints = append(ints, int32(v))
	}

	rows, err := r.db.Query(ctx, `
SELECT student_id, benefits, medical_notes, health_group, allergies, activities, updated_at
FROM students_medical WHERE student_id = ANY($1)`, ints)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int]*models.StudentMedical, len(ids))
	for rows.Next() {
		var m models.StudentMedical
		if err := rows.Scan(&m.StudentID, &m.Benefits, &m.MedicalNotes, &m.HealthGroup, &m.Allergies, &m.Activities, &m.UpdatedAt); err != nil {
			return nil, err
		}
		out[m.StudentID] = &m
	}
	return out, rows.Err()
}
