package info

import (
	"context"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/storage"
	"github.com/ARUMANDESU/uniclubs-club-service/pkg/logger"
	clubv1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/club"
	"log/slog"
)

var (
	ErrFailedToBeginTx   = errors.New("failed to begin transaction")
	ErrClubNotExists     = errors.New("club does not exists")
	ErrUserNotClubMember = errors.New("user is not a club member")
)

type Service struct {
	log     *slog.Logger
	storage Storage
}

type Storage interface {
	GetClubByID(ctx context.Context, clubID int64) (*domain.Club, error)
	GetMemberByID(ctx context.Context, clubID, userID int64) (*domain.User, error)
	GetUserClubsByID(ctx context.Context, userID int64) ([]*domain.Club, error)
	GetUserRoles(ctx context.Context, clubID, userID int64) (roles []*domain.Role, isOwner bool, err error)
	GetBanRecord(ctx context.Context, clubID, userID int64) (*domain.BanRecord, error)
	HaveUserJoinRequest(ctx context.Context, clubID, userID int64) (bool, error)
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
	ListMembershipRequests(ctx context.Context, clubID int64, filters domain.Filters) (
		[]*domain.User,
		*domain.Metadata,
		error,
	)
	ListBannedUsers(ctx context.Context, clubID int64, query string, filters domain.Filters) (
		[]*domain.BanRecord,
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
			return nil, fmt.Errorf("%s: %w", op, ErrClubNotExists)
		}
		log.Error("failed to get club by ID", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return club, nil
}

func (s Service) GetMemberByID(ctx context.Context, clubID, userID int64) (*domain.User, error) {
	const op = "services.info.GetMemberByID"
	log := s.log.With(slog.String("op", op))

	member, err := s.storage.GetMemberByID(ctx, clubID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotClubMember) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrMemberNotFound)
		}
		log.Error("failed to get member by ID", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return member, nil
}

func (s Service) GetUserRoles(ctx context.Context, clubID, userID int64) (roles []*domain.Role, isOwner bool, err error) {
	const op = "services.info.GetUser"
	log := s.log.With(slog.String("op", op))

	roles, isOwner, err = s.storage.GetUserRoles(ctx, clubID, userID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotClubMember):
			return nil, false, fmt.Errorf("%s: %w", op, domain.ErrUserNotClubMember)
		default:
			log.Error("failed to get user roles", logger.Err(err))
			return nil, false, fmt.Errorf("%s: %w", op, err)
		}
	}

	return roles, isOwner, nil
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

func (s Service) ListMembershipRequests(ctx context.Context, clubID int64, filters domain.Filters) ([]*domain.User, *domain.Metadata, error) {
	const op = "services.info.ListMembershipRequests"
	log := s.log.With(slog.String("op", op))

	users, metadata, err := s.storage.ListMembershipRequests(ctx, clubID, filters)
	if err != nil {
		log.Error("failed to get join requests of club", logger.Err(err))
		return nil, nil, err
	}

	return users, metadata, nil
}

func (s Service) GetJoinStatusOfUser(ctx context.Context, clubID, userID int64) (clubv1.JoinStatus, error) {
	const op = "services.info.GetJoinStatusOfUser"
	log := s.log.With(slog.String("op", op))

	member, err := s.storage.GetMemberByID(ctx, clubID, userID)
	if err != nil && !errors.Is(err, domain.ErrUserNotClubMember) {
		log.Error("failed to get club member", logger.Err(err))
		return clubv1.JoinStatus_NOT_MEMBER, fmt.Errorf("%s: %w", op, err)
	}
	if member != nil {
		return clubv1.JoinStatus_MEMBER, nil
	}

	haveUserJoinRequest, err := s.storage.HaveUserJoinRequest(ctx, clubID, userID)
	if err != nil {
		log.Error("failed to get info about user join request")
		return clubv1.JoinStatus_NOT_MEMBER, fmt.Errorf("%s: %w", op, err)
	}
	if haveUserJoinRequest {
		return clubv1.JoinStatus_PENDING, nil
	}

	banRecord, err := s.storage.GetBanRecord(ctx, clubID, userID)
	if err != nil && !errors.Is(err, storage.ErrBanRecordNotExists) {
		log.Error("failed to get ban record", logger.Err(err))
		return clubv1.JoinStatus_NOT_MEMBER, err
	}
	if banRecord != nil {
		return clubv1.JoinStatus_BANNED, nil
	}

	return clubv1.JoinStatus_NOT_MEMBER, nil
}

func (s Service) ListBannedUsers(ctx context.Context, clubID int64, query string, filters domain.Filters) ([]*domain.BanRecord, *domain.Metadata, error) {
	const op = "services.info.ListBannedUsers"
	log := s.log.With(slog.String("op", op))

	club, err := s.storage.GetClubByID(ctx, clubID)
	if err != nil {
		if errors.Is(err, storage.ErrClubNotExists) {
			log.Error("club does not exists", logger.Err(err))
			return nil, nil, fmt.Errorf("%s: %w", op, ErrClubNotExists)
		}
		log.Error("failed to get club by ID", logger.Err(err))
		return nil, nil, fmt.Errorf("%s: %w", op, err)
	}

	if club == nil {
		return nil, nil, fmt.Errorf("%s: %w", op, ErrClubNotExists)
	}

	bannedUsers, metadata, err := s.storage.ListBannedUsers(ctx, clubID, query, filters)
	if err != nil {
		log.Error("failed to get banned users", logger.Err(err))
		return nil, nil, err
	}

	return bannedUsers, metadata, nil
}
