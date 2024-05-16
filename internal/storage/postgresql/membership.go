package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/storage"
	"github.com/jackc/pgx/v5/pgconn"
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
		VALUES($1, $2, 0, 1, $3)
		RETURNING id, name, permissions, position, color`

	var role domain.Role

	args := []any{
		dto.ClubID, dto.Name, dto.Color,
	}

	err = tx.QueryRowContext(ctx, query, args...).Scan(&role.ID, &role.Name, &role.Permissions, &role.Position, &role.Color)
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

	// action -> Transaction[actions proverka if everything OK ] -> Commit
	// rollback -> Transaction[actions proverka if something went wrong] -> Rollback

	_, err = tx.ExecContext(ctx, `DELETE FROM users_roles WHERE role_id = $1`, roleID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to delete members of the role: %w", op, err)
	}

	var rolePos int

	err = tx.QueryRowContext(ctx, `DELETE FROM roles WHERE id = $1 AND club_id = $2 AND name != 'member' RETURNING position`, roleID, clubID).Scan(&rolePos)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			tx.Rollback()
			return storage.ErrClubOrRoleNotExists
		}
		tx.Rollback()
		return fmt.Errorf("%s: failed to delete role: %w", op, err)
	}

	_, err = tx.ExecContext(ctx, `UPDATE roles SET position = position - 1 WHERE club_id = $1 and name != 'member' AND position > $2`, clubID, rolePos)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to decrement upper roles: %w", op, err)
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

	err := s.DB.QueryRowContext(ctx, query, roleID, clubID).Scan(&role.ID, &role.Name, &role.Permissions, &role.Position, &role.Color, &role.UpdatedAt)
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
		strconv.FormatUint(role.Permissions, 10), role.Position,
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

func (s *Storage) ChangeRolesPosition(ctx context.Context, clubID int64, dto []*dtos.ChangeRolesPositionDTO) error {
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
		SET position = $3
		WHERE id = $1 AND club_id = $2 AND name != 'member'
	`
	for _, role := range dto {
		var isMemberRole bool
		err := tx.QueryRowContext(ctx, `SELECT name='member' FROM roles WHERE id = $1 AND club_id = $2`, role.RoleID, clubID).Scan(&isMemberRole)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				tx.Rollback()
				return domain.ErrClubOrRoleNotExists
			}
			tx.Rollback()
			return fmt.Errorf("%s: failed to get role: %w", op, err)
		}

		if isMemberRole {
			tx.Rollback()
			return domain.ErrCannotEditRoleMember
		}

		_, err = tx.ExecContext(ctx, query, role.RoleID, clubID, role.Position)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("%s: failed to change roles position: %w", op, err)
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
		WHERE club_id = $1
	`

	roles := []*domain.Role{}

	rows, err := s.DB.QueryContext(ctx, query, clubID)
	defer rows.Close()

	for rows.Next() {
		var role domain.Role
		err = rows.Scan(&role.ID, &role.Name, &role.Permissions, &role.Position, &role.Color, &role.UpdatedAt)
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

func (s *Storage) AddRoleMembers(ctx context.Context, clubID, roleID int64, usersID []int64) error {
	const op = "storage.postgresql.AddRoleMembers"

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

	addMemberQuery := `
		INSERT INTO users_roles(role_id, user_id)
		    values ($1, $2)
	`

	checkClubMemberQuery := `
		SELECT cu.user_id
		FROM clubs_users cu 
		WHERE cu.user_id = $1 AND cu.club_id = $2
	`

	for _, userID := range usersID {
		err = tx.QueryRowContext(ctx, checkClubMemberQuery, userID, clubID).Scan(&userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				tx.Rollback()
				return fmt.Errorf("%d: %w", userID, storage.ErrUserNotClubMember)
			}
			tx.Rollback()
			return fmt.Errorf("%s: failed to get a club member: %w", op, err)
		}

		_, err := tx.ExecContext(ctx, addMemberQuery, roleID, userID)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == "23505" {
					tx.Rollback()
					return fmt.Errorf("%d: %w", userID, domain.ErrUserAlreadyRoleMember)
				}
			}
			tx.Rollback()
			return fmt.Errorf("%s: failed to add new role members: %w", op, err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: transaction commit failed: %w", op, err)
	}

	return nil
}

func (s *Storage) RemoveMemberFromClub(ctx context.Context, clubID, userID int64) error {
	const op = "storage.postgresql.AddRoleMembers"

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

	result, err := tx.ExecContext(ctx, `DELETE FROM clubs_users WHERE user_id = $1 AND club_id = $2`, userID, clubID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to delete member from the club: %w", op, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to get rows affected from delete: %w", op, err)
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("%s: no rows updated, club or member may not exist: %w", op, domain.ErrMemberNotFound)
	}

	result, err = tx.ExecContext(ctx, `DELETE FROM users_roles ur USING roles r  WHERE ur.role_id = r.id AND ur.user_id = $1 AND r.club_id = $2`, userID, clubID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to remove roles membership: %w", op, err)
	}
	rowsAffected, err = result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to get rows affected from delete: %w", op, err)
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("%s: no rows deleted, club or member or roles may not exist: %w", op, domain.ErrMemberNotFound)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: transaction commit failed: %w", op, err)
	}

	return nil
}

func (s *Storage) RemoveRoleMembers(ctx context.Context, clubID, roleID int64, usersID []int64) error {
	const op = "storage.postgresql.RemoveRoleMembers"

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

	var isMemberRole bool
	err = tx.QueryRowContext(ctx, `SELECT name='member' FROM roles WHERE id = $1 AND club_id = $2`, roleID, clubID).Scan(&isMemberRole)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			tx.Rollback()
			return domain.ErrClubOrRoleNotExists
		}
		tx.Rollback()
		return fmt.Errorf("%s: failed to get role: %w", op, err)
	}

	if isMemberRole {
		tx.Rollback()
		return domain.ErrCannotEditRoleMember
	}

	removeRoleMember := `
		DELETE FROM users_roles ur
		       USING roles r
		       WHERE ur.user_id = $1 AND ur.role_id = $2 AND ur.role_id = r.id AND r.name != 'member'
	`

	checkClubMemberQuery := `
		SELECT cu.user_id
		FROM clubs_users cu 
		WHERE cu.user_id = $1 AND cu.club_id = $2
	`

	for _, userID := range usersID {
		err = tx.QueryRowContext(ctx, checkClubMemberQuery, userID, clubID).Scan(&userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				tx.Rollback()
				return fmt.Errorf("%d: %w", userID, storage.ErrUserNotClubMember)
			}
			tx.Rollback()
			return fmt.Errorf("%s: failed to get a club member: %w", op, err)
		}

		_, err := tx.ExecContext(ctx, removeRoleMember, userID, roleID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("%s: failed to remove role members: %w", op, err)
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: transaction commit failed: %w", op, err)
	}

	return nil
}

func (s *Storage) BanMember(ctx context.Context, dto dtos.BanMemberDTO) error {
	const op = "storage.postgresql.BanMember"

	query := `INSERT INTO bans(user_id, club_id, admin_id, reason, banned_at) VALUES ($1, $2, $3, $4, current_timestamp)`

	result, err := s.DB.ExecContext(ctx, query, dto.UserID, dto.ClubID, dto.AdminID, dto.Reason)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return fmt.Errorf("%d: %w", dto.UserID, domain.ErrUserAlreadyBanned)
			}
		}
		return fmt.Errorf("%s: failed to ban member : %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected from inserting into bans: %w", op, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%s: no rows inserted into bans", op)
	}

	return nil
}

func (s *Storage) UnbanUser(ctx context.Context, dto dtos.UnbanUserDTO) error {
	const op = "storage.postgresql.UnbanUser"

	query := `DELETE FROM bans WHERE user_id = $1 AND club_id = $2`

	result, err := s.DB.ExecContext(ctx, query, dto.UserID, dto.ClubID)
	if err != nil {
		return fmt.Errorf("%s: failed to unban member : %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: failed to get rows affected from deleting from bans: %w", op, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%s: no rows deleted from bans: %w", op, domain.ErrUserNotBanned)
	}

	return nil
}

func (s *Storage) GetBanRecord(ctx context.Context, clubID, userID int64) (*domain.BanRecord, error) {
	const op = "storage.postgresql.GetBanRecord"

	query := `
		SELECT id, user_id, club_id, admin_id, reason, banned_at
		FROM bans
		WHERE user_id = $1 AND club_id = $2
	`

	var banRecord domain.BanRecord

	err := s.DB.QueryRowContext(ctx, query, userID, clubID).Scan(&banRecord.ID, &banRecord.User.ID, &banRecord.ClubID, &banRecord.Admin.ID, &banRecord.Reason, &banRecord.BannedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrBanRecordNotExists
		}

		return nil, fmt.Errorf("%s: failed to get a ban record: %w", op, err)
	}

	return &banRecord, nil
}

func (s *Storage) DeleteClubByID(ctx context.Context, clubID int64) error {
	const op = "storage.postgresql.DeleteClubByID"

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	_, err = tx.ExecContext(ctx, `DELETE FROM clubs_users WHERE club_id = $1`, clubID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to delete club members: %w", op, err)
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM users_roles WHERE role_id IN (SELECT id FROM roles WHERE club_id = $1)`, clubID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to delete user roles: %w", op, err)
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM roles WHERE club_id = $1`, clubID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to delete club roles: %w", op, err)
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM join_club_requests WHERE club_id = $1`, clubID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to delete club join requests: %w", op, err)
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM bans WHERE club_id = $1`, clubID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to delete club bans: %w", op, err)
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM clubs WHERE id = $1`, clubID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to delete club: %w", op, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("%s: transaction commit failed: %w", op, err)
	}

	return nil
}
