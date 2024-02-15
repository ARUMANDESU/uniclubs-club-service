package info

import (
	"context"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/storage"
	"github.com/ARUMANDESU/uniclubs-club-service/pkg/logger"
	"log/slog"
)

var (
	ErrFailedToBeginTx = errors.New("failed to begin transaction")
	ErrClubNotExists   = errors.New("club does not exists")
)

type Service struct {
	log     *slog.Logger
	storage Storage
}

type Storage interface {
	GetClubByID(ctx context.Context, clubID int64) (*domain.Club, error)
	GetUserClubsByID(ctx context.Context, userID int64) ([]*domain.Club, error)
	ListClubs(
		ctx context.Context,
		query string,
		clubType []string,
		filters domain.Filters,
	) ([]*domain.Club, *domain.Metadata, error)
	ListNotApprovedClubs(
		ctx context.Context,
		query string,
		clubType []string,
		filters domain.Filters,
	) ([]*domain.ClubUser, *domain.Metadata, error)
	ListClubMembers(ctx context.Context, clubID int64, filters domain.Filters) (
		[]*domain.User,
		*domain.Metadata,
		error,
	)
}

func New(log *slog.Logger, storage Storage) *Service {
	return &Service{
		log:     log,
		storage: storage,
	}
}

func (s Service) GetClub(ctx context.Context, clubID int64) (*domain.Club, error) {
	const op = "services.info.GetClub"
	log := s.log.With(slog.String("op", op))

	club, err := s.storage.GetClubByID(ctx, clubID)
	if err != nil {
		if errors.Is(err, storage.ErrClubNotExists) {
			log.Error("club does not exists", logger.Err(err))
			return nil, fmt.Errorf("%s: %w", op, ErrClubNotExists)
		}
		log.Error("failed to get club by ID", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return club, nil
}

func (s Service) ListClub(ctx context.Context, query string, clubTypes []string, filters domain.Filters) ([]*domain.Club, *domain.Metadata, error) {
	const op = "services.info.ListClub"
	log := s.log.With(slog.String("op", op))

	clubs, metadata, err := s.storage.ListClubs(ctx, query, clubTypes, filters)
	if err != nil {
		log.Error("failed to get clubs", logger.Err(err))
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	return clubs, metadata, nil

}

func (s Service) ListNotActivatedClubs(ctx context.Context, query string, clubType []string, filters domain.Filters) ([]*domain.ClubUser, *domain.Metadata, error) {
	const op = "services.info.ListNotActivatedClubs"
	log := s.log.With(slog.String("op", op))

	clubsUsers, metadata, err := s.storage.ListNotApprovedClubs(ctx, query, clubType, filters)
	if err != nil {
		log.Error("failed to get clubs with users that not approved yet", logger.Err(err))
		return nil, nil, err
	}

	return clubsUsers, metadata, nil
}

func (s Service) GetUserClubs(ctx context.Context, userID int64) ([]*domain.Club, error) {
	const op = "services.info.GetUserClubs"
	log := s.log.With(slog.String("op", op))

	clubs, err := s.storage.GetUserClubsByID(ctx, userID)
	if err != nil {
		log.Error("failed to get clubs", logger.Err(err))
		return nil, err
	}

	return clubs, nil
}

func (s Service) ListClubMembers(ctx context.Context, clubID int64, filters domain.Filters) ([]*domain.User, *domain.Metadata, error) {
	const op = "services.info.ListClubMembers"
	log := s.log.With(slog.String("op", op))

	members, metadata, err := s.storage.ListClubMembers(ctx, clubID, filters)
	if err != nil {
		log.Error("failed to get members of club", logger.Err(err))
		return nil, nil, err
	}

	return members, metadata, nil
}
