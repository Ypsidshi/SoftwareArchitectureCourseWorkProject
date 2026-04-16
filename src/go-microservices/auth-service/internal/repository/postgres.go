package repository

import (
	"context"
	"database/sql"
	"errors"

	"coursework/auth-service/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrEmailExists = errors.New("email already exists")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *UserRepository) CreateUser(ctx context.Context, email, fullName, role, passwordHash string) (domain.User, error) {
	const q = `
INSERT INTO auth.users (email, full_name, role, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING id, email, full_name, role, password_hash, created_at`

	var user domain.User
	err := r.db.QueryRowContext(ctx, q, email, fullName, role, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, ErrEmailExists
		}
		return domain.User{}, err
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	const q = `
SELECT id, email, full_name, role, password_hash, created_at
FROM auth.users
WHERE email = $1`

	var user domain.User
	err := r.db.QueryRowContext(ctx, q, email).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.Role,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}
