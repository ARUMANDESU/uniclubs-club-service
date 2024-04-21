package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	DB *sql.DB
}

func New(databaseDSN string) (*Storage, error) {
	const op = "storage.postgresql.New"

	db, err := sql.Open("pgx", databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{DB: db}, nil
}

func (s *Storage) GetUserRoles(ctx context.Context, clubID, userID int64) (roles []*domain.Role, isOwner bool, err error) {
	const op = "storage.postgresql.GetUserRoles"

	query := `
		SELECT c.owner_id = cu.user_id AS is_owner, r.id,
		       r.name, r.permissions, r.position, r.color
		FROM clubs_users cu
		JOIN clubs c ON cu.club_id = c.id  
		JOIN users_roles ur ON cu.user_id= ur.user_id 
		JOIN roles r ON ur.role_id = r.id 
		WHERE c.id = $1 AND cu.user_id = $2 AND r.club_id = $1;

	`

	rows, err := s.DB.QueryContext(ctx, query, clubID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, storage.ErrUserNotClubMember
		}
		return nil, false, fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close()

	for rows.Next() {
		var role domain.Role
		err = rows.Scan(&isOwner, &role.ID, &role.Name, &role.Permissions, &role.Position, &role.Color)
		if err != nil {
			return nil, false, fmt.Errorf("%s: %w", op, err)
		}
		roles = append(roles, &role)
	}
	if err = rows.Err(); err != nil {
		return nil, false, fmt.Errorf("%s: %w", op, err)
	}

	if len(roles) == 0 {
		return nil, false, fmt.Errorf("%s: %w", op, storage.ErrUserNotClubMember)
	}

	return roles, isOwner, nil
}

func (s *Storage) GetClubRoles(ctx context.Context, clubID int64) ([]*domain.Role, error) {
	const op = "storage.postgresql.GetClubRoles"

	query := `
		SELECT id, name, position, permissions, color
    	FROM roles 
		WHERE club_id = $1
    `

	rows, err := s.DB.QueryContext(ctx, query, clubID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var roles []*domain.Role

	for rows.Next() {
		var role domain.Role

		err := rows.Scan(&role.ID, &role.Name, &role.Position, &role.Permissions, &role.Color)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		roles = append(roles, &role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: iterating and scanning roles: %w", op, err)
	}

	return roles, nil

}
