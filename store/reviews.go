package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

var ErrReviewNotFound = errors.New("review not found")
var ErrReviewAlreadyExists = errors.New("review already exists")

type Review struct {
	ID               string  `json:"id"`
	UserID           string  `json:"user_id"`
	CafeID           string  `json:"cafe_id"`
	MemberName       string  `json:"member_name"`
	CafeName         string  `json:"cafe_name"`
	Rating           int     `json:"rating"`
	DrinkRating      *int    `json:"drink_rating"`
	DessertRating    *int    `json:"dessert_rating"`
	AtmosphereRating *int    `json:"atmosphere_rating"`
	ServiceRating    *int    `json:"service_rating"`
	Comment          *string `json:"comment"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// GetReviewsByCafeID retrieves reviews of a cafe.
func (s *Store) GetReviewsByCafeID(
	ctx context.Context,
	cafeID string,
) ([]Review, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT
			id,
			user_id,
			cafe_id,
			rating,
			drink_rating,
			dessert_rating,
			atmosphere_rating,
			service_rating,
			comment,
			created_at::text,
			updated_at::text
		FROM reviews
		WHERE cafe_id = $1
		ORDER BY created_at DESC
	`, cafeID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]Review, 0)

	for rows.Next() {
		var review Review

		err := rows.Scan(
			&review.ID,
			&review.UserID,
			&review.CafeID,
			&review.Rating,
			&review.DrinkRating,
			&review.DessertRating,
			&review.AtmosphereRating,
			&review.ServiceRating,
			&review.Comment,
			&review.CreatedAt,
			&review.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reviews, nil
}

// GetReviewByID retrieves a review by ID.
func (s *Store) GetReviewByID(
	ctx context.Context,
	id string,
) (Review, error) {
	var review Review

	err := s.DB.QueryRow(ctx, `
		SELECT
			id,
			user_id,
			cafe_id,
			rating,
			drink_rating,
			dessert_rating,
			atmosphere_rating,
			service_rating,
			comment,
			created_at::text,
			updated_at::text
		FROM reviews
		WHERE id = $1
	`, id).Scan(
		&review.ID,
		&review.UserID,
		&review.CafeID,
		&review.Rating,
		&review.DrinkRating,
		&review.DessertRating,
		&review.AtmosphereRating,
		&review.ServiceRating,
		&review.Comment,
		&review.CreatedAt,
		&review.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Review{}, ErrReviewNotFound
	}

	if err != nil {
		return Review{}, err
	}

	return review, nil
}

// UpdateReview updates a member's own review.
func (s *Store) UpdateReview(
	ctx context.Context,
	review Review,
) error {
	tag, err := s.DB.Exec(ctx, `
		UPDATE reviews
		SET
			rating = $1,
			drink_rating = $2,
			dessert_rating = $3,
			atmosphere_rating = $4,
			service_rating = $5,
			comment = $6,
			updated_at = NOW()
		WHERE id = $7 AND user_id = $8
	`,
		review.Rating,
		review.DrinkRating,
		review.DessertRating,
		review.AtmosphereRating,
		review.ServiceRating,
		review.Comment,
		review.ID,
		review.UserID,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrReviewNotFound
	}

	return nil
}

// DeleteReview deletes a member's own review.
func (s *Store) DeleteReview(
	ctx context.Context,
	reviewID string,
	userID string,
) error {
	tag, err := s.DB.Exec(ctx, `
		DELETE FROM reviews
		WHERE id = $1 AND user_id = $2
	`, reviewID, userID)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrReviewNotFound
	}

	return nil
}

// GetAllReviews retrieves all reviews for admin moderation.
func (s *Store) GetAllReviews(
	ctx context.Context,
) ([]Review, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT
			r.id,
			r.user_id,
			r.cafe_id,
			COALESCE(NULLIF(u.full_name, ''), u.email, 'สมาชิก') AS member_name,
			COALESCE(NULLIF(c.name, ''), 'ไม่ระบุคาเฟ่') AS cafe_name,
			r.rating,
			r.drink_rating,
			r.dessert_rating,
			r.atmosphere_rating,
			r.service_rating,
			r.comment,
			r.created_at::text,
			r.updated_at::text
		FROM reviews r
		LEFT JOIN users u ON u.id = r.user_id
		LEFT JOIN cafes c ON c.id = r.cafe_id
		ORDER BY r.created_at DESC
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]Review, 0)

	for rows.Next() {
		var review Review

		err := rows.Scan(
			&review.ID,
			&review.UserID,
			&review.CafeID,
			&review.MemberName,
			&review.CafeName,
			&review.Rating,
			&review.DrinkRating,
			&review.DessertRating,
			&review.AtmosphereRating,
			&review.ServiceRating,
			&review.Comment,
			&review.CreatedAt,
			&review.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reviews, nil
}

// AdminDeleteReview deletes any review for admin moderation.
func (s *Store) AdminDeleteReview(
	ctx context.Context,
	reviewID string,
) error {
	tag, err := s.DB.Exec(ctx, `
		DELETE FROM reviews
		WHERE id = $1
	`, reviewID)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrReviewNotFound
	}

	return nil
}
