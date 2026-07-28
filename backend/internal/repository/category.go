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

type CategoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

func (r *CategoryRepository) CreateCategory(
	ctx context.Context,
	input domain.CreateCategoryInput,
) (*domain.Category, error) {
	const query = `
		INSERT INTO categories (user_id, name, category_type, color, icon)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id::text,
			name,
			category_type,
			color,
			icon,
			created_at,
			updated_at
	`

	var category domain.Category
	err := r.pool.QueryRow(
		ctx,
		query,
		input.UserID,
		input.Name,
		input.CategoryType,
		input.Color,
		input.Icon,
	).Scan(
		&category.ID,
		&category.Name,
		&category.CategoryType,
		&category.Color,
		&category.Icon,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrCategoryAlreadyExists
		}
		return nil, fmt.Errorf("create category: %w", err)
	}
	return &category, nil
}

func (r *CategoryRepository) ListCategories(
	ctx context.Context,
	filter domain.ListCategoriesFilter,
) ([]domain.Category, error) {
	query := `
		SELECT
			id::text,
			name,
			category_type,
			color,
			icon,
			created_at,
			updated_at
		FROM categories
		WHERE user_id = $1
	`
	args := []any{filter.UserID}

	if filter.CategoryType != nil {
		query += ` AND category_type = $2`
		args = append(args, *filter.CategoryType)
	}

	query += ` ORDER BY category_type ASC, name ASC, id ASC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	categories := make([]domain.Category, 0)
	for rows.Next() {
		var category domain.Category
		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.CategoryType,
			&category.Color,
			&category.Icon,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}
	return categories, nil
}

func (r *CategoryRepository) UpdateCategory(
	ctx context.Context,
	input domain.UpdateCategoryInput,
) (*domain.Category, error) {
	const query = `
		UPDATE categories
		SET
			name = COALESCE($3, name),
			category_type = COALESCE($4, category_type),
			color = COALESCE($5, color),
			icon = COALESCE($6, icon),
			updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING
			id::text,
			name,
			category_type,
			color,
			icon,
			created_at,
			updated_at
	`

	var category domain.Category
	err := r.pool.QueryRow(
		ctx,
		query,
		input.CategoryID,
		input.UserID,
		input.Name,
		input.CategoryType,
		input.Color,
		input.Icon,
	).Scan(
		&category.ID,
		&category.Name,
		&category.CategoryType,
		&category.Color,
		&category.Icon,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrCategoryNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrCategoryAlreadyExists
		}
		return nil, fmt.Errorf("update category: %w", err)
	}
	return &category, nil
}

func (r *CategoryRepository) DeleteCategory(ctx context.Context, userID, categoryID string) error {
	const query = `
		DELETE FROM categories
		WHERE id = $1 AND user_id = $2
	`

	tag, err := r.pool.Exec(ctx, query, categoryID, userID)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrCategoryNotFound
	}
	return nil
}
