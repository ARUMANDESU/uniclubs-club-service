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
		"INSERT INTO clubs (name, description, type) VALUES ($1, $2, $3) RETURNING id",
		dto.Name,
		dto.Description,
		dto.ClubType,
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

	// Delete create club request
	_, err = tx.ExecContext(ctx, `DELETE FROM create_club_requests WHERE club_id = $1`, clubID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to delete create club request: %w", op, err)
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
