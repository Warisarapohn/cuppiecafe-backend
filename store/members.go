package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// ErrMemberNotFound is returned when a member does not exist.
var ErrMemberNotFound = errors.New("member not found")

// Member represents a Cuppie user.
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

// CreateMember inserts a new member into the database.
func (s *Store) CreateMember(ctx context.Context, member Member) error {
	_, err := s.DB.Exec(
		ctx,
		`INSERT INTO users (full_name, email, password_hash, role)
		 VALUES ($1, $2, $3, 'member')`,
		member.FullName,
		member.Email,
		member.PasswordHash,
	)

	return err
}

// GetMemberByEmail retrieves an active member by email.
func (s *Store) GetMemberByEmail(
	ctx context.Context,
	email string,
) (Member, error) {
	var member Member

	err := s.DB.QueryRow(
		ctx,
		`SELECT id, full_name, email, password_hash, role,
		        avatar_url, bio, is_active
		 FROM users
		 WHERE LOWER(email) = LOWER($1)
		   AND is_active = TRUE`,
		email,
	).Scan(
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
	if err != nil {
		return Member{}, err
	}

	return member, nil
}

// GetMemberByID retrieves an active member by ID.
func (s *Store) GetMemberByID(
	ctx context.Context,
	id string,
) (Member, error) {
	var member Member

	err := s.DB.QueryRow(
		ctx,
		`SELECT id, full_name, email, password_hash, role,
		        avatar_url, bio, is_active
		 FROM users
		 WHERE id = $1
		   AND is_active = TRUE`,
		id,
	).Scan(
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
	if err != nil {
		return Member{}, err
	}

	return member, nil
}

// GetAllMembers retrieves all member accounts for admin.
func (s *Store) GetAllMembers(ctx context.Context) ([]Member, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, full_name, email, role,
		       avatar_url, bio, is_active
		FROM users
		WHERE role = 'member'
		ORDER BY full_name ASC
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

// GetMemberByIDForAdmin retrieves a member account for admin.
func (s *Store) GetMemberByIDForAdmin(
	ctx context.Context,
	id string,
) (Member, error) {
	var member Member

	err := s.DB.QueryRow(ctx, `
		SELECT id, full_name, email, role,
		       avatar_url, bio, is_active
		FROM users
		WHERE id = $1 AND role = 'member'
	`, id).Scan(
		&member.ID,
		&member.FullName,
		&member.Email,
		&member.Role,
		&member.AvatarURL,
		&member.Bio,
		&member.IsActive,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrMemberNotFound
	}
	if err != nil {
		return Member{}, err
	}

	return member, nil
}

// DeactivateMember deactivates a member account.
func (s *Store) DeactivateMember(ctx context.Context, id string) error {
	tag, err := s.DB.Exec(ctx, `
		UPDATE users
		SET is_active = FALSE,
		    updated_at = NOW()
		WHERE id = $1
		  AND role = 'member'
		  AND is_active = TRUE
	`, id)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrMemberNotFound
	}

	return nil
}
