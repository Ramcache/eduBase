package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eduBase/internal/models"
)

type ConsentsRepository interface {
	Upsert(ctx context.Context, studentID int, pd bool, pdDate *time.Time, photo bool, phDate *time.Time, net bool, netDate *time.Time) error
	Get(ctx context.Context, studentID int) (*models.StudentConsents, error)
	BulkByStudentIDs(ctx context.Context, ids []int) (map[int]*models.StudentConsents, error)
}

type consentsRepo struct{ db *pgxpool.Pool }

func NewConsentsRepo(db *pgxpool.Pool) ConsentsRepository { return &consentsRepo{db: db} }

func (r *consentsRepo) Upsert(ctx context.Context, studentID int, pd bool, pdDate *time.Time, photo bool, phDate *time.Time, net bool, netDate *time.Time) error {
	_, err := r.db.Exec(ctx, `
INSERT INTO students_consents(student_id,consent_data_processing,consent_data_processing_date,
  consent_photo_publication,consent_photo_publication_date,
  consent_internet_access,consent_internet_access_date)
VALUES($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (student_id) DO UPDATE SET
  consent_data_processing=EXCLUDED.consent_data_processing,
  consent_data_processing_date=EXCLUDED.consent_data_processing_date,
  consent_photo_publication=EXCLUDED.consent_photo_publication,
  consent_photo_publication_date=EXCLUDED.consent_photo_publication_date,
  consent_internet_access=EXCLUDED.consent_internet_access,
  consent_internet_access_date=EXCLUDED.consent_internet_access_date,
  updated_at=now()`, studentID, pd, pdDate, photo, phDate, net, netDate)
	return err
}

func (r *consentsRepo) Get(ctx context.Context, studentID int) (*models.StudentConsents, error) {
	row := r.db.QueryRow(ctx, `
SELECT student_id, consent_data_processing, consent_data_processing_date,
       consent_photo_publication, consent_photo_publication_date,
       consent_internet_access, consent_internet_access_date, updated_at
FROM students_consents WHERE student_id=$1`, studentID)

	var m models.StudentConsents
	if err := row.Scan(&m.StudentID, &m.ConsentPD, &m.ConsentPDDate, &m.ConsentPhoto, &m.ConsentPhotoDate, &m.ConsentInternet, &m.ConsentInternetDate, &m.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *consentsRepo) BulkByStudentIDs(ctx context.Context, ids []int) (map[int]*models.StudentConsents, error) {
	if len(ids) == 0 {
		return map[int]*models.StudentConsents{}, nil
	}
	ints := make([]int32, 0, len(ids))
	for _, v := range ids {
		ints = append(ints, int32(v))
	}

	rows, err := r.db.Query(ctx, `
SELECT student_id, consent_data_processing, consent_data_processing_date,
       consent_photo_publication, consent_photo_publication_date,
       consent_internet_access, consent_internet_access_date, updated_at
FROM students_consents WHERE student_id = ANY($1)`, ints)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int]*models.StudentConsents, len(ids))
	for rows.Next() {
		var m models.StudentConsents
		if err := rows.Scan(&m.StudentID, &m.ConsentPD, &m.ConsentPDDate, &m.ConsentPhoto, &m.ConsentPhotoDate, &m.ConsentInternet, &m.ConsentInternetDate, &m.UpdatedAt); err != nil {
			return nil, err
		}
		out[m.StudentID] = &m
	}
	return out, rows.Err()
}
