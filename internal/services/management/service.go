package management

import (
	"context"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-club-service/pkg/logger"
	"log/slog"
)

var (
	ErrFailedToBeginTx = errors.New("failed to begin transaction")
)

type Service struct {
	log     *slog.Logger
	storage Storage
}

type Storage interface {
	SaveClub(ctx context.Context, dto dtos.CreateClubDTO) error
	ApproveClub(ctx context.Context, clubID int64) error
	RejectClub(ctx context.Context, clubID int64) error
}

func New(log *slog.Logger, storage Storage) *Service {
	return &Service{
		log:     log,
		storage: storage,
	}
}

func (s Service) CreateClub(ctx context.Context, dto dtos.CreateClubDTO) error {
	const op = "services.management.CreateClub"
	log := s.log.With(slog.String("op", op))

	err := s.storage.SaveClub(ctx, dto)
	if err != nil {
		log.Error("failed to create club", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s Service) ApproveClub(ctx context.Context, clubID int64) error {
	const op = "services.management.ApproveClub"
	log := s.log.With(slog.String("op", op))

	err := s.storage.ApproveClub(ctx, clubID)
	if err != nil {
		log.Error("failed to approve club", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s Service) RejectClub(ctx context.Context, clubID int64) error {
	const op = "services.management.RejectClub"
	log := s.log.With(slog.String("op", op))

	err := s.storage.RejectClub(ctx, clubID)
	if err != nil {
		log.Error("failed to reject club", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s Service) GetClub(ctx context.Context, clubID int64) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) ListClub(ctx context.Context, query, clubType string, filters domain.Filters) ([]*domain.Club, domain.Metadata, error) {
	//TODO implement me
	panic("implement me")
}

func (s Service) ListNotActivatedClubs(ctx context.Context, query, clubType string, filters domain.Filters) ([]*domain.ClubUser, domain.Metadata, error) {
	//TODO implement me
	panic("implement me")
}

func (s Service) CreateJoinRequest(ctx context.Context, userID, clubID int64) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) ApproveMembership(ctx context.Context, userID, memberID, clubID int64) error {
	//TODO implement me
	panic("implement me")
}

func (s Service) RejectMembership(ctx context.Context, userID, memberID, clubID int64) error {
	//TODO implement me
	panic("implement me")
}
