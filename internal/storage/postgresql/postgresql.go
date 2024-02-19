package postgresql

import (
	"context"
	"database/sql"
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

func (s *Storage) GetUserRoles(ctx context.Context, clubID, userID int64) (roles []domain.Role, isOwner bool, err error) {
	const op = "storage.postgresql.GetUserRoles"

	query := `
		SELECT c.owner_id = cu.user_id as is_owner, r.id, r.name, r.permissions, r.position, r.color
		FROM clubs_users cu
		LEFT JOIN clubs c ON c.id = cu.club_id
		JOIN users_roles ur ON ur.user_id = cu.user_id
		JOIN roles r ON r.id = ur.role_id
		WHERE cu.club_id = $1 and cu.user_id = $2
	`

	rows, err := s.DB.QueryContext(ctx, query, clubID, userID)
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close()

	for rows.Next() {
		var role domain.Role
		err = rows.Scan(&isOwner, &role.ID, &role.Name, &role.Permissions.PermissionsHex, &role.Position, &role.Color)
		if err != nil {
			return nil, false, fmt.Errorf("%s: %w", op, err)
		}
		roles = append(roles, role)
	}
	if err = rows.Err(); err != nil {
		return nil, false, fmt.Errorf("%s: %w", op, err)
	}

	if len(roles) == 0 {
		return nil, false, fmt.Errorf("%s: %w", op, storage.ErrUserNotClubMember)
	}

	return roles, isOwner, nil
}
