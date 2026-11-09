package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// ErrCafeNotFound is returned when a cafe does not exist.
var ErrCafeNotFound = errors.New("cafe not found")

// Cafe represents a cafe in Cuppie.
type Cafe struct {
	ID           string   `json:"id"`
	FSQID        *string  `json:"fsq_id"`
	CategoryID   *string  `json:"category_id"`
	CategoryName *string  `json:"category_name"`
	Name         string   `json:"name"`
	Description  *string  `json:"description"`
	Address      *string  `json:"address"`
	PriceMin     *float64 `json:"price_min"`
	PriceMax     *float64 `json:"price_max"`
	OpeningHours *string  `json:"opening_hours"`
	ContactInfo  *string  `json:"contact_info"`
}

// GetAllCafes retrieves all active cafes.
func (s *Store) GetAllCafes(ctx context.Context) ([]Cafe, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT
			c.id,
			c.fsq_id,
			c.category_id,
			ca.name,
			c.name,
			c.description,
			c.address,
			c.price_min,
			c.price_max,
			c.opening_hours,
			c.contact_info
		FROM cafes c
		LEFT JOIN categories ca ON ca.id = c.category_id
		WHERE c.is_active = TRUE
		ORDER BY c.created_at DESC
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cafes := make([]Cafe, 0)

	for rows.Next() {
		var cafe Cafe

		if err := rows.Scan(
			&cafe.ID,
			&cafe.FSQID,
			&cafe.CategoryID,
			&cafe.CategoryName,
			&cafe.Name,
			&cafe.Description,
			&cafe.Address,
			&cafe.PriceMin,
			&cafe.PriceMax,
			&cafe.OpeningHours,
			&cafe.ContactInfo,
		); err != nil {
			return nil, err
		}

		cafes = append(cafes, cafe)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cafes, nil
}

// GetCafeByID retrieves an active cafe by ID.
func (s *Store) GetCafeByID(ctx context.Context, id string) (Cafe, error) {
	var cafe Cafe

	err := s.DB.QueryRow(ctx, `
		SELECT
			c.id,
			c.fsq_id,
			c.category_id,
			ca.name,
			c.name,
			c.description,
			c.address,
			c.price_min,
			c.price_max,
			c.opening_hours,
			c.contact_info
		FROM cafes c
		LEFT JOIN categories ca ON ca.id = c.category_id
		WHERE c.id = $1 AND c.is_active = TRUE
	`, id).Scan(
		&cafe.ID,
		&cafe.FSQID,
		&cafe.CategoryID,
		&cafe.CategoryName,
		&cafe.Name,
		&cafe.Description,
		&cafe.Address,
		&cafe.PriceMin,
		&cafe.PriceMax,
		&cafe.OpeningHours,
		&cafe.ContactInfo,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Cafe{}, ErrCafeNotFound
	}

	if err != nil {
		return Cafe{}, err
	}

	return cafe, nil
}

// GetOrCreateCafeByFSQID finds an existing cafe or creates a new one.
func (s *Store) GetOrCreateCafeByFSQID(
	ctx context.Context,
	place FoursquarePlace,
) (Cafe, error) {
	if place.FSQID == "" {
		return Cafe{}, errors.New("Foursquare ID is required")
	}

	var id string

	err := s.DB.QueryRow(ctx, `
		INSERT INTO cafes (
			fsq_id,
			name,
			address,
			is_active
		)
		VALUES ($1, $2, $3, TRUE)
		ON CONFLICT (fsq_id) WHERE fsq_id IS NOT NULL
		DO UPDATE SET
			name = EXCLUDED.name,
			address = COALESCE(
				NULLIF(EXCLUDED.address, ''),
				cafes.address
			),
			is_active = TRUE,
			updated_at = NOW()
		RETURNING id
	`,
		place.FSQID,
		place.Name,
		place.Address,
	).Scan(&id)

	if err != nil {
		return Cafe{}, err
	}

	return s.GetCafeByID(ctx, id)
}

// CreateCafe creates a new cafe.
func (s *Store) CreateCafe(ctx context.Context, cafe Cafe) (Cafe, error) {
	var id string

	err := s.DB.QueryRow(ctx, `
		INSERT INTO cafes (
			category_id,
			name,
			description,
			address,
			price_min,
			price_max,
			opening_hours,
			contact_info,
			is_active
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, TRUE)
		RETURNING id
	`,
		cafe.CategoryID,
		cafe.Name,
		cafe.Description,
		cafe.Address,
		cafe.PriceMin,
		cafe.PriceMax,
		cafe.OpeningHours,
		cafe.ContactInfo,
	).Scan(&id)

	if err != nil {
		return Cafe{}, err
	}

	return s.GetCafeByID(ctx, id)
}

// UpdateCafe updates an active cafe.
func (s *Store) UpdateCafe(
	ctx context.Context,
	id string,
	cafe Cafe,
) (Cafe, error) {
	result, err := s.DB.Exec(ctx, `
		UPDATE cafes
		SET
			category_id = $1,
			name = $2,
			description = $3,
			address = $4,
			price_min = $5,
			price_max = $6,
			opening_hours = $7,
			contact_info = $8,
			updated_at = NOW()
		WHERE id = $9 AND is_active = TRUE
	`,
		cafe.CategoryID,
		cafe.Name,
		cafe.Description,
		cafe.Address,
		cafe.PriceMin,
		cafe.PriceMax,
		cafe.OpeningHours,
		cafe.ContactInfo,
		id,
	)

	if err != nil {
		return Cafe{}, err
	}

	if result.RowsAffected() == 0 {
		return Cafe{}, ErrCafeNotFound
	}

	return s.GetCafeByID(ctx, id)
}

// DeactivateCafe deactivates a cafe without deleting its data.
func (s *Store) DeactivateCafe(ctx context.Context, id string) error {
	result, err := s.DB.Exec(ctx, `
		UPDATE cafes
		SET
			is_active = FALSE,
			updated_at = NOW()
		WHERE id = $1 AND is_active = TRUE
	`, id)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrCafeNotFound
	}

	return nil
}
