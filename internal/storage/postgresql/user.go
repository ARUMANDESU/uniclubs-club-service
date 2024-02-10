package postgresql

import (
	"context"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/storage"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s Storage) SaveUser(ctx context.Context, user *domain.User) error {
	const op = "storage.postgresql.SaveUser"

	stmt, err := s.DB.Prepare(`
		INSERT INTO users(id, email, first_name, last_name, barcode, avatar_url)
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
		user.FirstName,
		user.LastName,
		user.Barcode,
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

func (s Storage) GetUserByID(ctx context.Context, userID int64) (user *domain.User, err error) {
	//TODO implement me
	panic("implement me")
}

func (s Storage) UpdateUser(ctx context.Context, user *domain.User) error {
	//TODO implement me
	panic("implement me")
}

func (s Storage) DeleteUserByID(ctx context.Context, userID int64) error {
	//TODO implement me
	panic("implement me")
}
