package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eduBase/internal/models"
)

type DocumentsRepository interface {
	Upsert(ctx context.Context, studentID int, snils string, ser, num, cert *string) error
	Get(ctx context.Context, studentID int) (*models.StudentDocuments, error)
	GetRaw(ctx context.Context, studentID int) (snils *string, ser *string, num *string, cert *string, err error)
	BulkByStudentIDs(ctx context.Context, ids []int) (map[int]*models.StudentDocuments, error)
}

type documentsRepo struct{ db *pgxpool.Pool }

func NewDocumentsRepo(db *pgxpool.Pool) DocumentsRepository { return &documentsRepo{db: db} }

func (r *documentsRepo) Upsert(ctx context.Context, studentID int, snils string, ser, num, cert *string) error {
	_, err := r.db.Exec(ctx, `
INSERT INTO students_documents(student_id,snils,passport_series,passport_number,birth_certificate)
VALUES($1,$2,$3,$4,$5)
ON CONFLICT (student_id) DO UPDATE
   SET snils=EXCLUDED.snils,
       passport_series=EXCLUDED.passport_series,
       passport_number=EXCLUDED.passport_number,
       birth_certificate=EXCLUDED.birth_certificate,
       updated_at=now()`,
		studentID, snils, ser, num, cert)
	return err
}

func (r *documentsRepo) Get(ctx context.Context, studentID int) (*models.StudentDocuments, error) {
	row := r.db.QueryRow(ctx, `
SELECT student_id, snils, passport_series, passport_number, birth_certificate, updated_at
FROM students_documents WHERE student_id=$1`, studentID)

	var m models.StudentDocuments
	if err := row.Scan(&m.StudentID, &m.SNILS, &m.PassportSeries, &m.PassportNumber, &m.BirthCertificate, &m.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *documentsRepo) GetRaw(ctx context.Context, studentID int) (*string, *string, *string, *string, error) {
	row := r.db.QueryRow(ctx, `
SELECT snils, passport_series, passport_number, birth_certificate
FROM students_documents WHERE student_id=$1`, studentID)

	var snils, ser, num, cert *string
	err := row.Scan(&snils, &ser, &num, &cert)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, nil, nil, nil
	}
	return snils, ser, num, cert, err
}

func (r *documentsRepo) BulkByStudentIDs(ctx context.Context, ids []int) (map[int]*models.StudentDocuments, error) {
	if len(ids) == 0 {
		return map[int]*models.StudentDocuments{}, nil
	}
	ints := make([]int32, 0, len(ids))
	for _, v := range ids {
		ints = append(ints, int32(v))
	}

	rows, err := r.db.Query(ctx, `
SELECT student_id, snils, passport_series, passport_number, birth_certificate, updated_at
FROM students_documents
WHERE student_id = ANY($1)`, ints)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int]*models.StudentDocuments, len(ids))
	for rows.Next() {
		var m models.StudentDocuments
		if err := rows.Scan(&m.StudentID, &m.SNILS, &m.PassportSeries, &m.PassportNumber, &m.BirthCertificate, &m.UpdatedAt); err != nil {
			return nil, err
		}
		out[m.StudentID] = &m
	}
	return out, rows.Err()
}
