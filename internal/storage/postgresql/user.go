package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/storage"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Storage) SaveUser(ctx context.Context, user *domain.User) error {
	const op = "storage.postgresql.SaveUser"

	stmt, err := s.DB.Prepare(`
		INSERT INTO users(id, email, barcode, first_name, last_name, avatar_url)
		values($1, $2, $3, $4, $5, $6)
		returning id;
	`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	args := []any{
		user.ID,
		user.Email,
		user.Barcode,
		user.FirstName,
		user.LastName,
		user.AvatarURL,
	}

	result := stmt.QueryRowContext(ctx, args...)

	err = result.Scan(&user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return fmt.Errorf("%s: %w", op, storage.ErrUserExists)
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetUserByID(ctx context.Context, userID int64) (*domain.User, error) {
	const op = "storage.postgresql.GetUserByID"

	stmt, err := s.DB.Prepare(`
		SELECT id, email, barcode, first_name, last_name, avatar_url
		FROM users
		WHERE id = $1;
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	result := stmt.QueryRowContext(ctx, userID)

	var user domain.User

	err = result.Scan(
		&user.ID, &user.Email, &user.Barcode,
		&user.FirstName, &user.LastName, &user.AvatarURL,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotExists)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) UpdateUser(ctx context.Context, user *domain.User) error {
	const op = "storage.postgresql.UpdateUser"

	stmt, err := s.DB.Prepare(`
		UPDATE users
		SET email = $2,barcode = $3, first_name = $4, last_name = $5 ,avatar_url = $6
		WHERE id = $1;
	`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	args := []any{
		user.ID,
		user.Email,
		user.Barcode,
		user.FirstName,
		user.LastName,
		user.AvatarURL,
	}
	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrUserNotExists)
	}

	return nil
}

func (s *Storage) DeleteUserByID(ctx context.Context, userID int64) error {
	const op = "storage.postgresql.DeleteUserByID"

	stmt, err := s.DB.Prepare(`DELETE FROM users WHERE id = $1;`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrUserNotExists)
	}

	return nil
}
