package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

var ErrFavoriteNotFound = errors.New("favorite not found")
var ErrFavoriteAlreadyExists = errors.New("favorite already exists")

type Favorite struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	CafeID    string `json:"cafe_id"`
	CafeName  string `json:"cafe_name"`
	CreatedAt string `json:"created_at"`
}

func (s *Store) GetFavoritesByUserID(
	ctx context.Context,
	userID string,
) ([]Favorite, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT f.id, f.user_id, f.cafe_id, c.name, f.created_at::text
		FROM favorites f
		JOIN cafes c ON c.id = f.cafe_id
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	favorites := make([]Favorite, 0)

	for rows.Next() {
		var favorite Favorite

		if err := rows.Scan(
			&favorite.ID,
			&favorite.UserID,
			&favorite.CafeID,
			&favorite.CafeName,
			&favorite.CreatedAt,
		); err != nil {
			return nil, err
		}

		favorites = append(favorites, favorite)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return favorites, nil
}

func (s *Store) AddFavorite(
	ctx context.Context,
	userID string,
	cafeID string,
) error {
	_, err := s.DB.Exec(ctx, `
		INSERT INTO favorites (user_id, cafe_id)
		VALUES ($1, $2)
	`, userID, cafeID)

	return err
}

func (s *Store) DeleteFavorite(
	ctx context.Context,
	userID string,
	cafeID string,
) error {
	tag, err := s.DB.Exec(ctx, `
		DELETE FROM favorites
		WHERE user_id = $1 AND cafe_id = $2
	`, userID, cafeID)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrFavoriteNotFound
	}

	return nil
}

// Keep pgx imported for checking missing records when needed.
var _ = pgx.ErrNoRows
