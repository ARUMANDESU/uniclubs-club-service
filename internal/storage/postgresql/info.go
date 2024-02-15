package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/storage"
	"time"
)

func (s *Storage) GetClubByID(ctx context.Context, clubID int64) (*domain.Club, error) {
	const op = "storage.postgresql.GetClubByID"

	clubQuery := `
        SELECT id, name, description, type, logo_url, banner_url, created_at, COUNT(user_id) as member_count
        FROM clubs
        LEFT JOIN clubs_users ON clubs.id = clubs_users.club_id
        WHERE clubs.id = $1 AND approved
        GROUP BY clubs.id;
    `

	var club domain.Club
	err := s.DB.QueryRowContext(ctx, clubQuery, clubID).Scan(
		&club.ID,
		&club.Name,
		&club.Description,
		&club.ClubType,
		&club.LogoURL,
		&club.BannerURL,
		&club.CreatedAt,
		&club.NumOFMembers,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrClubNotExists)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rolesQuery := `
        SELECT id, name
        FROM roles
        WHERE club_id = $1;
    `
	rolesRows, err := s.DB.QueryContext(ctx, rolesQuery, clubID)
	if err != nil {
		return nil, fmt.Errorf("%s: querying roles: %w", op, err)
	}
	defer rolesRows.Close()

	var roles []domain.Role
	for rolesRows.Next() {
		var r domain.Role
		if err := rolesRows.Scan(&r.ID, &r.Name); err != nil {
			return nil, fmt.Errorf("%s: scanning roles: %w", op, err)
		}
		roles = append(roles, r)
	}
	if err := rolesRows.Err(); err != nil {
		return nil, fmt.Errorf("%s: iterating roles: %w", op, err)
	}

	// Query to get permissions for each role.
	for i, role := range roles {
		permissionsQuery := `
            SELECT permissions.name
            FROM permissions
            JOIN roles_permissions ON permissions.id = roles_permissions.permission_id
            WHERE roles_permissions.role_id = $1;
        `
		permissionsRows, err := s.DB.QueryContext(ctx, permissionsQuery, role.ID)
		if err != nil {
			return nil, fmt.Errorf("%s: querying permissions: %w", op, err)
		}

		var permissions []string
		for permissionsRows.Next() {
			var p string
			if err := permissionsRows.Scan(&p); err != nil {
				permissionsRows.Close()
				return nil, fmt.Errorf("%s: scanning permissions: %w", op, err)
			}
			permissions = append(permissions, p)
		}
		if err := permissionsRows.Err(); err != nil {
			permissionsRows.Close()
			return nil, fmt.Errorf("%s: iterating permissions: %w", op, err)
		}
		permissionsRows.Close()

		// Assign the permissions to the role
		roles[i].Permissions = permissions

	}

	// Assign the roles to the club and return.
	club.Roles = roles

	return &club, nil
}

func (s *Storage) ListClubs(
	ctx context.Context,
	query string,
	clubTypes []string,
	filters domain.Filters,
) ([]*domain.Club, *domain.Metadata, error) {
	const op = "storage.postgresql.ListClubs"

	stmt, err := s.DB.Prepare(`
		SELECT count(*) OVER(), c.id, c.name, c.description, c.type, c.logo_url, c.banner_url, c.created_at, COUNT(cu.user_id) as member_count
		FROM clubs c
		LEFT JOIN clubs_users cu ON c.id = cu.club_id
		WHERE  
		    ( (STRPOS(LOWER(c.name), LOWER($1)) > 0 OR $1 = '') OR
			(STRPOS(LOWER(c.description), LOWER($1)) > 0 OR $1 = '') )
			AND	(type = ANY($2) OR $2::text[] IS NULL)
			AND c.approved
		GROUP BY c.id
		ORDER BY c.id
		LIMIT $3 OFFSET $4;
	`)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	args := []any{query, clubTypes, filters.Limit(), filters.Offset()}

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close()

	var totalRecords int32
	clubs := []*domain.Club{}

	for rows.Next() {
		var club domain.Club

		err := rows.Scan(
			&totalRecords, &club.ID, &club.Name,
			&club.Description, &club.ClubType, &club.LogoURL,
			&club.BannerURL, &club.CreatedAt, &club.NumOFMembers,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", op, err)
		}

		clubs = append(clubs, &club)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	metadata := domain.CalculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return clubs, &metadata, nil

}

func (s *Storage) ListNotApprovedClubs(
	ctx context.Context,
	query string,
	clubTypes []string,
	filters domain.Filters,
) ([]*domain.ClubUser, *domain.Metadata, error) {
	const op = "storage.postgresql.ListNotApprovedClubs"

	stmt, err := s.DB.Prepare(`
		SELECT count(*) OVER(), c.id, c.name,
		       c.description, c.type, c.logo_url,
		       c.banner_url, c.created_at, COUNT(ccr.user_id) as member_count,
		       u.id, u.email, u.barcode, u.first_name, u.last_name, u.avatar_url
		FROM clubs c
		JOIN create_club_requests ccr ON c.id = ccr.club_id
		JOIN users u ON u.id = ccr.user_id
		WHERE  
		    ( (STRPOS(LOWER(c.name), LOWER($1)) > 0 OR $1 = '') OR
			(STRPOS(LOWER(c.description), LOWER($1)) > 0 OR $1 = '') )
			AND	(type = ANY($2) OR $2::text[] IS NULL)
			AND NOT c.approved
		GROUP BY c.id, u.id
		ORDER BY c.id
		LIMIT $3 OFFSET $4;
	`)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	args := []any{query, clubTypes, filters.Limit(), filters.Offset()}

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	defer rows.Close()

	var totalRecords int32
	clubsUsers := []*domain.ClubUser{}

	for rows.Next() {
		var clubUser domain.ClubUser

		err := rows.Scan(
			&totalRecords, &clubUser.Club.ID, &clubUser.Club.Name,
			&clubUser.Club.Description, &clubUser.Club.ClubType, &clubUser.Club.LogoURL,
			&clubUser.Club.BannerURL, &clubUser.Club.CreatedAt, &clubUser.Club.NumOFMembers,
			&clubUser.User.ID, &clubUser.User.Email, &clubUser.User.Barcode,
			&clubUser.User.FirstName, &clubUser.User.LastName, &clubUser.User.AvatarURL,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", op, err)
		}

		clubsUsers = append(clubsUsers, &clubUser)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	metadata := domain.CalculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return clubsUsers, &metadata, nil
}

func (s *Storage) GetUserClubsByID(ctx context.Context, userID int64) ([]*domain.Club, error) {
	const op = "storage.postgresql.GetUserClubsByID"

	stmt, err := s.DB.Prepare(`
		SELECT c.id, c.name, c.description, c.type, c.logo_url,
		       c.banner_url, c.created_at, (SELECT COUNT(cu.club_id)FROM clubs_users cu WHERE cu.club_id = c.id GROUP BY cu.club_id) as member_count
		FROM users u
		LEFT JOIN clubs_users cu2 ON u.id = cu2.user_id
		JOIN clubs c ON c.id = cu2.club_id
		WHERE u.id = $1
		GROUP BY c.id, u.id;
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var clubs []*domain.Club

	for rows.Next() {
		var club domain.Club

		err = rows.Scan(
			&club.ID, &club.Name,
			&club.Description, &club.ClubType, &club.LogoURL,
			&club.BannerURL, &club.CreatedAt, &club.NumOFMembers,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		clubs = append(clubs, &club)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return clubs, nil

}

func (s *Storage) ListClubMembers(ctx context.Context, clubID int64, filters domain.Filters) (
	[]*domain.User,
	*domain.Metadata,
	error,
) {
	const op = "storage.postgresql.GetUserClubsByID"

	stmt, err := s.DB.Prepare(`
		SELECT count(*) OVER(), u.id, u.email, u.barcode, u.first_name, u.last_name, u.avatar_url, r.name as role
		FROM clubs c 
		JOIN clubs_users cu ON c.id = cu.club_id
		JOIN users u ON cu.user_id = u.id
		JOIN roles r ON r.id = cu.role_id AND r.club_id = c.id
		WHERE c.id = $1
		LIMIT $2 OFFSET $3;
	`)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := stmt.QueryContext(ctx, clubID, filters.Limit(), filters.Offset())
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var totalRecords int32
	var users []*domain.User

	for rows.Next() {
		var user domain.User

		err = rows.Scan(&totalRecords, &user.ID, &user.Email, &user.Barcode, &user.FirstName, &user.LastName, &user.AvatarURL, &user.Role)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", op, err)
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	metadata := domain.CalculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return users, &metadata, nil
}
