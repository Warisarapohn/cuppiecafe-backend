package store

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	DB *pgxpool.Pool
}

func NewDB() (*pgxpool.Pool, error) {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid DATABASE_URL: %w", err)
	}

	// ปิด Prepared Statement Cache
	// เพื่อรองรับ Supabase Transaction Pooler
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	config.MaxConns = 5
	config.MinConns = 0
	config.MaxConnLifetime = time.Hour

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{DB: db}
}

// CreateReview creates a new review.
func (s *Store) CreateReview(ctx context.Context, review Review) error {
	_, err := s.DB.Exec(ctx, `
		INSERT INTO reviews (
			user_id,
			cafe_id,
			rating,
			drink_rating,
			dessert_rating,
			atmosphere_rating,
			service_rating,
			comment
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		review.UserID,
		review.CafeID,
		review.Rating,
		review.DrinkRating,
		review.DessertRating,
		review.AtmosphereRating,
		review.ServiceRating,
		review.Comment,
	)

	return err
}
