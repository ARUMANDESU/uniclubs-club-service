package postgresql

import (
	"context"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain/dtos"
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
