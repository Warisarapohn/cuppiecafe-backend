package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

var ErrCategoryNotFound = errors.New("category not found")

type Category struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (s *Store) GetAllCategories(ctx context.Context) ([]Category, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, name, description
		FROM categories
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]Category, 0)

	for rows.Next() {
		var category Category

		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
		)
		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *Store) GetCategoryByID(
	ctx context.Context,
	id string,
) (Category, error) {
	var category Category

	err := s.DB.QueryRow(ctx, `
		SELECT id, name, description
		FROM categories
		WHERE id = $1
	`, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Category{}, ErrCategoryNotFound
	}

	if err != nil {
		return Category{}, err
	}

	return category, nil
}
