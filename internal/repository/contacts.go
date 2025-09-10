package repository

import (
	"context"

	"eduBase/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ContactsRepository interface {
	Add(ctx context.Context, studentID int, fullName, phone, relation string) (int, error)
	Delete(ctx context.Context, id int) error
	ListByStudent(ctx context.Context, studentID int) ([]models.EmergencyContact, error)
	BulkByStudentIDs(ctx context.Context, ids []int) (map[int][]models.EmergencyContact, error)
	DeleteByStudent(ctx context.Context, studentID int) error
}

type contactsRepo struct{ db *pgxpool.Pool }

func NewContactsRepo(db *pgxpool.Pool) ContactsRepository {
	return &contactsRepo{db: db}
}

func (r *contactsRepo) Add(ctx context.Context, studentID int, fullName, phone, relation string) (int, error) {
	var id int
	err := r.db.QueryRow(ctx, `
INSERT INTO students_emergency_contacts(student_id,full_name,phone,relation)
VALUES($1,$2,$3,$4) RETURNING id`, studentID, fullName, phone, relation).Scan(&id)
	return id, err
}

func (r *contactsRepo) Delete(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM students_emergency_contacts WHERE id=$1`, id)
	return err
}

func (r *contactsRepo) ListByStudent(ctx context.Context, studentID int) ([]models.EmergencyContact, error) {
	rows, err := r.db.Query(ctx, `
SELECT id,student_id,full_name,phone,relation,created_at
FROM students_emergency_contacts WHERE student_id=$1 ORDER BY id`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.EmergencyContact
	for rows.Next() {
		var e models.EmergencyContact
		if err := rows.Scan(&e.ID, &e.StudentID, &e.FullName, &e.Phone, &e.Relation, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *contactsRepo) BulkByStudentIDs(ctx context.Context, ids []int) (map[int][]models.EmergencyContact, error) {
	if len(ids) == 0 {
		return map[int][]models.EmergencyContact{}, nil
	}
	ints := make([]int32, 0, len(ids))
	for _, v := range ids {
		ints = append(ints, int32(v))
	}

	rows, err := r.db.Query(ctx, `
SELECT id, student_id, full_name, phone, relation, created_at
FROM students_emergency_contacts
WHERE student_id = ANY($1)
ORDER BY student_id, id`, ints)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[int][]models.EmergencyContact, len(ids))
	for rows.Next() {
		var e models.EmergencyContact
		if err := rows.Scan(&e.ID, &e.StudentID, &e.FullName, &e.Phone, &e.Relation, &e.CreatedAt); err != nil {
			return nil, err
		}
		out[e.StudentID] = append(out[e.StudentID], e)
	}
	return out, rows.Err()
}

func (r *contactsRepo) DeleteByStudent(ctx context.Context, studentID int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM students_emergency_contacts WHERE student_id=$1`, studentID)
	return err
}
