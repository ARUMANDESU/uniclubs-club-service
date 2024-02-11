package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/storage"
	"github.com/lib/pq"
	"time"
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
	err = tx.QueryRowContext(ctx, `INSERT INTO roles(club_id, name) VALUES ($1, $2) returning id`, clubID, "president").Scan(&roleID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to insert president role: %w", op, err)
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO clubs_users(user_id, club_id, role_id) VALUES ($1, $2, $3)`, userID, clubID, roleID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: failed to insert to clubs_users: %w", op, err)
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

func (s *Storage) GetClubByID(ctx context.Context, clubID int64) (*domain.Club, error) {
	const op = "storage.postgresql.GetClubByID"

	stmt, err := s.DB.Prepare(`
		SELECT c.id, c.name, c.description, c.type, c.logo_url, c.banner_url, c.created_at, COUNT(cu.user_id) as member_count
		FROM clubs c
		LEFT JOIN clubs_users cu ON c.id = cu.club_id
		WHERE c.id = $1 and approved
		GROUP BY c.id;
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer stmt.Close()

	result := stmt.QueryRowContext(ctx, clubID)
	var club domain.Club

	err = result.Scan(
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

	return &club, nil
}

func (s *Storage) ListClubs(ctx context.Context, query string, clubTypes []string, filters domain.Filters) ([]*domain.Club, *domain.Metadata, error) {
	const op = "storage.postgresql.GetClubByID"

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

	args := []any{query, pq.Array(clubTypes), filters.Limit(), filters.Offset()}

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
