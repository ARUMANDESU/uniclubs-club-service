package accessControl

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
	ErrInsufficientRolePosition = errors.New("user's highest role position is less than target's")
	ErrUserNotClubMember        = errors.New("user is not club member")
	ErrTargetNotClubMember      = errors.New("target user is not club member")
)

type Service struct {
	log     *slog.Logger
	storage Storage
}

type Storage interface {
	GetUserRoles(ctx context.Context, clubID, userID int64) (roles []domain.Role, isOwner bool, err error)
}

func New(log *slog.Logger, storage Storage) *Service {
	return &Service{
		log:     log,
		storage: storage,
	}
}

func (s *Service) CanActOnMember(ctx context.Context, clubID, userID, targetID int64, permission uint64) (bool, error) {
	const op = "service.accessControl.CheckPermission"
	log := s.log.With(slog.String("op", op))

	userRoles, isUserOwner, err := s.storage.GetUserRoles(ctx, clubID, userID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotClubMember):
			log.Error("user is not club member", logger.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrUserNotClubMember)
		default:
			log.Error("failed to get user permissions", logger.Err(err))
			return false, fmt.Errorf("%s: %w", op, err)
		}

	}
	if isUserOwner {
		return true, nil
	}

	targetRoles, isTargetOwner, err := s.storage.GetUserRoles(ctx, clubID, targetID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotClubMember):
			log.Error("user is not club member", logger.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrTargetNotClubMember)
		default:
			log.Error("failed to get user permissions", logger.Err(err))
			return false, fmt.Errorf("%s: %w", op, err)
		}

	}
	if isTargetOwner {
		return false, nil
	}

	userHighestPos, err := domain.GetHighestRolePosition(userRoles)
	if err != nil {
		log.Error("failed to get user highest role position", logger.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}
	targetHighestPos, err := domain.GetHighestRolePosition(targetRoles)
	if err != nil {
		log.Error("failed to get target highest role position", logger.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if userHighestPos < targetHighestPos {
		log.Error(ErrInsufficientRolePosition.Error())
		return false, fmt.Errorf("%s: %w", op, ErrInsufficientRolePosition)
	}

	userPermissions := domain.AccumulatePermissions(userRoles)

	return domain.HasPermission(userPermissions, permission), nil
}

func (s *Service) CanHandleMembershipRequest(ctx context.Context, clubID, userID int64) (bool, error) {
	const op = "service.accessControl.CanHandleMembershipRequest"
	log := s.log.With(slog.String("op", op))

	userRoles, isUserOwner, err := s.storage.GetUserRoles(ctx, clubID, userID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotClubMember):
			log.Error("user is not club member", logger.Err(err))
			return false, fmt.Errorf("%s: %w", op, ErrUserNotClubMember)
		default:
			log.Error("failed to get user permissions", logger.Err(err))
			return false, fmt.Errorf("%s: %w", op, err)
		}
	}
	if isUserOwner {
		return true, nil
	}

	userPermissions := domain.AccumulatePermissions(userRoles)

	return domain.HasPermission(userPermissions, domain.ManageMembership), nil

}
