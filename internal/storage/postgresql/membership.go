package postgresql

import (
	"context"
	"fmt"
	"time"
)

func (s *Storage) InsertJoinRequest(ctx context.Context, userID, clubID int64) error {
	const op = "storage.postgresql.InsertJoinRequest"

	stmt, err := s.DB.Prepare(`INSERT INTO join_club_requests(user_id, club_id) VALUES ($1, $2)`)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	result, err := stmt.ExecContext(ctx, userID, clubID)
	if err != nil {
		return fmt.Errorf("%s: failed to execute query: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected from insert: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: no row inserted", op)
	}

	return nil
}

func (s *Storage) AddNewMember(ctx context.Context, clubID, userID int64) error {
	const op = "storage.postgresql.AddNewMember"

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

	deleteQuery := `
		DELETE FROM join_club_requests 
		WHERE club_id = $1 AND user_id = $2;
	`
	result, err := tx.ExecContext(ctx, deleteQuery, clubID, userID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to delete join request: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to get rows affected from delete: %w", op, err)
	}
	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("%s: no rows deleted from join requests", op)
	}

	result, err = tx.ExecContext(ctx, `INSERT INTO clubs_users(user_id, club_id) VALUES ($1, $2);`, userID, clubID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to insert to clubs_users: %w", op, err)
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to get rows affected from insert into clubs_users: %w", op, err)
	}
	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("%s: no rows inserted into clubs_users", op)
	}

	var roleID int
	err = tx.QueryRowContext(ctx, `SELECT id FROM roles WHERE club_id = $1 AND name = 'member';`, clubID).Scan(&roleID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to get member role id of club: %w", op, err)
	}

	result, err = tx.ExecContext(ctx, `INSERT INTO users_roles(user_id, role_id) VALUES ($1, $2);`, userID, roleID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to insert to users_roles: %w", op, err)
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to get rows affected from insert into users_roles: %w", op, err)
	}
	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("%s: no rows inserted into users_roles", op)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: transaction commit failed: %w", op, err)
	}

	return nil
}

func (s *Storage) DeleteJoinRequest(ctx context.Context, clubID, userID int64) error {
	const op = "storage.postgresql.DeleteJoinRequest"

	query := `DELETE FROM join_club_requests 
		WHERE club_id = $1 AND user_id = $2; 
	`
	result, err := s.DB.ExecContext(ctx, query, clubID, userID)
	if err != nil {
		return fmt.Errorf("%s: failed to delete join request: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected from delete: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: no rows updated", op)
	}

	return nil
}
