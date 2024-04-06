package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/storage"
	"strconv"
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

func (s *Storage) CreateRole(ctx context.Context, dto dtos.CreateRoleDTO) (*domain.Role, error) {
	const op = "storage.postgresql.CreateRole"

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}

	// Defer the rollback in case of any error.
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	_, err = tx.ExecContext(ctx, `UPDATE roles SET position = position + 1 WHERE club_id = $1 and name != 'member'`, dto.ClubID)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("%s: failed to increment roles position: %w", op, err)
	}

	query := `
		INSERT INTO roles(club_id, name, permissions, position, color) 
		VALUES($1, $2, $3, $4, $5)
		RETURNING id, name, permissions, position, color`

	var role domain.Role

	args := []any{
		dto.ClubID, dto.Name,
		dto.Permissions, dto.Position,
		dto.Color,
	}

	err = tx.QueryRowContext(ctx, query, args...).Scan(&role.ID, &role.Name, &role.Permissions.PermissionsHex, &role.Position, &role.Color)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("%s: failed to insert role: %w", op, err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("%s: transaction commit failed: %w", op, err)
	}

	return &role, nil
}

func (s *Storage) DeleteRoleByID(ctx context.Context, clubID, roleID int64) error {
	const op = "storage.postgresql.DeleteRoleByID"

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

	_, err = tx.ExecContext(ctx, `DELETE FROM users_roles WHERE role_id = $1`, roleID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to delete members of the role: %w", op, err)
	}

	result, err := tx.ExecContext(ctx, `DELETE FROM roles WHERE id = $1 AND club_id = $2`, roleID, clubID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to delete role: %w", op, err)
	}

	// Check the number of rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to get rows affected from delete: %w", op, err)
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("%s: no rows deleted, club or role may not exist: %w", op, storage.ErrClubOrRoleNotExists)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: transaction commit failed: %w", op, err)
	}

	return nil
}

func (s *Storage) GetRoleByID(ctx context.Context, clubID, roleID int64) (*domain.Role, error) {
	const op = "storage.postgresql.GetRoleByID"

	query := `
		SELECT id, name, permissions, position, color, updated_at
		FROM roles
		WHERE id = $1 AND club_id = $2
	`

	var role domain.Role

	err := s.DB.QueryRowContext(ctx, query, roleID, clubID).Scan(&role.ID, &role.Name, &role.Permissions.PermissionsHex, &role.Position, &role.Color, &role.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrClubOrRoleNotExists
		}

		return nil, fmt.Errorf("%s: failed to get a role: %w", op, err)
	}

	return &role, nil

}

func (s *Storage) UpdateRole(ctx context.Context, role *domain.Role) error {
	const op = "storage.postgresql.UpdateRole"

	query := `
		UPDATE roles 
		SET name = $3, permissions = $4, position = $5, color = $6
		WHERE id = $1 AND club_id = $2 AND updated_at = $7
		returning updated_at
	`

	args := []any{
		role.ID, role.ClubID, role.Name,
		strconv.FormatUint(role.Permissions.PermissionsHex, 10), role.Position,
		role.Color, role.UpdatedAt,
	}

	err := s.DB.QueryRowContext(ctx, query, args...).Scan(&role.UpdatedAt)
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

func (s *Storage) ChangeRolesPosition(ctx context.Context, dto []*dtos.ChangeRolesPositionDTO) error {
	const op = "storage.postgresql.ChangeRolesPosition"
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

	query := `
		UPDATE roles
		SET position = $2
		WHERE id = $1
	`
	for _, role := range dto {
		_, err := tx.ExecContext(ctx, query, role.RoleID, role.Position)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("%s: failed to delete role: %w", op, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: transaction commit failed: %w", op, err)
	}

	return nil
}

func (s *Storage) GetRolesOfClubByID(ctx context.Context, clubID int64) ([]*domain.Role, error) {
	const op = "storage.postgresql.GetRolesOfClubByID"

	query := `
		SELECT id, name, permissions, position, color, updated_at
		FROM roles
		WHERE club_id = $1 AND name != 'member'
	`

	roles := []*domain.Role{}

	rows, err := s.DB.QueryContext(ctx, query, clubID)
	defer rows.Close()

	for rows.Next() {
		var role domain.Role
		err = rows.Scan(&role.ID, &role.Name, &role.Permissions.PermissionsHex, &role.Position, &role.Color, &role.UpdatedAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, storage.ErrClubOrRoleNotExists
			}
			return nil, fmt.Errorf("%s: failed to get a role: %w", op, err)
		}

		roles = append(roles, &role)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return roles, nil
}
