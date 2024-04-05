package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/storage"
)

func (s *Storage) SaveClub(ctx context.Context, dto dtos.CreateClubDTO) error {
	const op = "storage.postgresql.SaveClub"

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}

	// Defer the rollback in case of any error.
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	var clubID int64
	// Insert into the clubs table and get id.
	err = tx.QueryRowContext(
		ctx,
		"INSERT INTO clubs (name, description, type, owner_id) VALUES ($1, $2, $3, $4) RETURNING id",
		dto.Name,
		dto.Description,
		dto.ClubType,
		dto.OwnerID,
	).Scan(&clubID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s, failed to insert into clubs and get club_id: %w", op, err)
	}

	// Insert into the requests_create_club table.
	_, err = tx.ExecContext(ctx, "INSERT INTO create_club_requests (club_id, user_id) VALUES ($1, $2)", clubID, dto.OwnerID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to insert create club request: %w", op, err)
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: transaction commit failed: %w", op, err)
	}

	return nil
}

func (s *Storage) ApproveClub(ctx context.Context, clubID int64) error {
	const op = "storage.postgresql.ApproveClub"

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}

	// Defer the rollback in case of any error.
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	var (
		userID int64
		roleID int
	)
	// Delete create club request
	err = tx.QueryRowContext(ctx, `DELETE FROM create_club_requests WHERE club_id = $1 RETURNING user_id`, clubID).Scan(&userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to delete create club request and get userID: %w", op, err)
	}

	// Update club approved to true
	result, err := tx.ExecContext(ctx, `UPDATE clubs SET approved = true WHERE id = $1 and not approved`, clubID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to update club approved to true: %w", op, err)
	}
	// Check the number of rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to get rows affected from update: %w", op, err)
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("%s: no rows updated, club may already be approved or does not exist", op)
	}

	// New president role
	err = tx.QueryRowContext(ctx, `INSERT INTO roles(club_id, name, permissions, position, color) VALUES ($1, $2, '0', 0, 8223868) returning id`, clubID, "member").Scan(&roleID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to insert president role: %w", op, err)
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO clubs_users(user_id, club_id) VALUES ($1, $2)`, userID, clubID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to insert to clubs_users: %w", op, err)
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO users_roles(user_id, role_id) VALUES ($1, $2)`, userID, roleID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to insert to users_roles: %w", op, err)
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: transaction commit failed: %w", op, err)
	}

	return nil
}

func (s *Storage) RejectClub(ctx context.Context, clubID int64) error {
	const op = "storage.postgresql.RejectClub"

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}

	// Defer the rollback in case of any error.
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	// Delete create club request
	_, err = tx.ExecContext(ctx, `DELETE FROM create_club_requests WHERE club_id = $1`, clubID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to delete create club request: %w", op, err)
	}

	// Delete club
	result, err := tx.ExecContext(ctx, `DELETE FROM clubs WHERE id = $1 and not approved`, clubID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to delete club: %w", op, err)
	}

	// Check the number of rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to get rows affected from delete: %w", op, err)
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("%s: no rows deleted, club may not exist", op)
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: transaction commit failed: %w", op, err)
	}

	return nil
}

func (s *Storage) UpdateClub(ctx context.Context, club *domain.Club) error {
	const op = "storage.postgresql.UpdateClub"

	query := `
        UPDATE clubs 
        SET name = $2, description = $3, type = $4,
            logo_url = $5, banner_url = $6, updated_at = current_timestamp
        WHERE id = $1 AND approved AND updated_at = $7
        returning updated_at
    `

	args := []any{
		club.ID, club.Name, club.Description,
		club.ClubType, club.LogoURL, club.BannerURL,
		club.UpdatedAt,
	}

	err := s.DB.QueryRowContext(ctx, query, args...).Scan(&club.UpdatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return storage.ErrEditConflict
		default:
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}
