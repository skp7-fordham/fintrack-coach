package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/skp7-fordham/fintrack-coach/backend/internal/domain"
)

type AuthRepository struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool}
}

func (r *AuthRepository) CreateUser(
	ctx context.Context,
	input domain.RegisterUserInput,
) (*domain.User, error) {
	const query = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id::text, email, is_demo, created_at, updated_at
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, input.Email, input.PasswordHash).Scan(
		&user.ID,
		&user.Email,
		&user.IsDemo,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}

func (r *AuthRepository) FindUserByEmail(
	ctx context.Context,
	normalizedEmail string,
) (*domain.UserWithPasswordHash, error) {
	const query = `
		SELECT id::text, email, password_hash, is_demo, created_at, updated_at
		FROM users
		WHERE LOWER(email) = $1
		LIMIT 1
	`

	var user domain.UserWithPasswordHash
	err := r.pool.QueryRow(ctx, query, normalizedEmail).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsDemo,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	return &user, nil
}
