package membership

import (
	"context"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/pkg/logger"
	"log/slog"
)

type Service struct {
	log     *slog.Logger
	storage Storage
}

type Storage interface {
	InsertJoinRequest(ctx context.Context, userID, clubID int64) error
	AddNewMember(ctx context.Context, clubID, userID int64) error
	DeleteJoinRequest(ctx context.Context, clubID, userID int64) error
}

func New(log *slog.Logger, storage Storage) *Service {
	return &Service{
		log:     log,
		storage: storage,
	}
}

func (s Service) CreateJoinRequest(ctx context.Context, userID, clubID int64) error {
	const op = "services.membership.CreateJoinRequest"
	log := s.log.With(slog.String("op", op))

	err := s.storage.InsertJoinRequest(ctx, userID, clubID)
	if err != nil {
		log.Error("failed to create new join request", logger.Err(err))
		return err
	}

	return nil
}

func (s Service) ApproveMembership(ctx context.Context, clubID, userID int64) error {
	const op = "services.membership.ApproveMembership"
	log := s.log.With(slog.String("op", op))

	err := s.storage.AddNewMember(ctx, clubID, userID)
	if err != nil {
		log.Error("failed to add not member to club", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s Service) RejectMembership(ctx context.Context, clubID, userID int64) error {
	const op = "services.membership.RejectMembership"
	log := s.log.With(slog.String("op", op))

	err := s.storage.DeleteJoinRequest(ctx, clubID, userID)
	if err != nil {
		log.Error("failed to delete join request", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
