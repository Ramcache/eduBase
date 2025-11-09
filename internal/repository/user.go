package repository

import (
	"context"
	"errors"

	"eduBase/internal/models"
	"github.com/jackc/pgx/v5"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	db *pgx.Conn
}

func NewUserRepository(db *pgx.Conn) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	row := r.db.QueryRow(ctx, `SELECT id, email, password, role, created_at FROM users WHERE email=$1`, email)
	var u models.User
	if err := row.Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, u *models.User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (email, password, role)
		VALUES ($1, $2, $3)
	`, u.Email, u.Password, u.Role)
	return err
}

func (r *UserRepository) DB() *pgx.Conn {
	return r.db
}
