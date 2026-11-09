package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

var ErrMemberNotFound = errors.New("member not found")

type Member struct {
	ID           string  `json:"id"`
	FullName     string  `json:"full_name"`
	Email        string  `json:"email"`
	PasswordHash string  `json:"-"`
	Role         string  `json:"role"`
	AvatarURL    *string `json:"avatar_url"`
	Bio          *string `json:"bio"`
	IsActive     bool    `json:"is_active"`
}

// CreateMember creates a new member account.
func (s *Store) CreateMember(ctx context.Context, member Member) error {
	_, err := s.DB.Exec(ctx, `
		INSERT INTO users (
			full_name,
			email,
			password_hash,
			role,
			avatar_url,
			bio
		)
		VALUES ($1, $2, $3, 'member', $4, $5)
	`,
		member.FullName,
		member.Email,
		member.PasswordHash,
		member.AvatarURL,
		member.Bio,
	)

	return err
}

// GetMemberByEmail finds a user by email,
// including inactive accounts.
// Login checks IsActive after verifying the password.
func (s *Store) GetMemberByEmail(ctx context.Context, email string) (Member, error) {
	var member Member

	err := s.DB.QueryRow(ctx, `
		SELECT
			id,
			full_name,
			email,
			password_hash,
			role,
			avatar_url,
			bio,
			is_active
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`, email).Scan(
		&member.ID,
		&member.FullName,
		&member.Email,
		&member.PasswordHash,
		&member.Role,
		&member.AvatarURL,
		&member.Bio,
		&member.IsActive,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrMemberNotFound
	}

	return member, err
}

// GetMemberByID finds an active member by ID.
func (s *Store) GetMemberByID(ctx context.Context, id string) (Member, error) {
	var member Member

	err := s.DB.QueryRow(ctx, `
		SELECT
			id,
			full_name,
			email,
			password_hash,
			role,
			avatar_url,
			bio,
			is_active
		FROM users
		WHERE id = $1
		  AND is_active = TRUE
	`, id).Scan(
		&member.ID,
		&member.FullName,
		&member.Email,
		&member.PasswordHash,
		&member.Role,
		&member.AvatarURL,
		&member.Bio,
		&member.IsActive,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrMemberNotFound
	}

	return member, err
}

// GetAllMembers returns all user accounts,
// including admins and inactive accounts.
func (s *Store) GetAllMembers(ctx context.Context) ([]Member, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT
			id,
			full_name,
			email,
			password_hash,
			role,
			avatar_url,
			bio,
			is_active
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]Member, 0)

	for rows.Next() {
		var member Member

		err := rows.Scan(
			&member.ID,
			&member.FullName,
			&member.Email,
			&member.PasswordHash,
			&member.Role,
			&member.AvatarURL,
			&member.Bio,
			&member.IsActive,
		)
		if err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

// GetMemberByIDForAdmin finds any user account by ID.
func (s *Store) GetMemberByIDForAdmin(ctx context.Context, id string) (Member, error) {
	var member Member

	err := s.DB.QueryRow(ctx, `
		SELECT
			id,
			full_name,
			email,
			password_hash,
			role,
			avatar_url,
			bio,
			is_active
		FROM users
		WHERE id = $1
	`, id).Scan(
		&member.ID,
		&member.FullName,
		&member.Email,
		&member.PasswordHash,
		&member.Role,
		&member.AvatarURL,
		&member.Bio,
		&member.IsActive,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrMemberNotFound
	}

	return member, err
}

// DeactivateMember deactivates a user account.
func (s *Store) DeactivateMember(ctx context.Context, id string) error {
	result, err := s.DB.Exec(ctx, `
		UPDATE users
		SET
			is_active = FALSE,
			updated_at = NOW()
		WHERE id = $1
		  AND is_active = TRUE
	`, id)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrMemberNotFound
	}

	return nil
}
