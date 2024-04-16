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
	ErrClubOrRoleNotExists      = errors.New("club or role does not exists")
)

type Service struct {
	log     *slog.Logger
	storage Storage
}

type Storage interface {
	GetUserRoles(ctx context.Context, clubID, userID int64) (roles []*domain.Role, isOwner bool, err error)
	GetClubRoles(ctx context.Context, clubID int64) ([]*domain.Role, error)
	GetRoleByID(ctx context.Context, clubID, roleID int64) (*domain.Role, error)
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

func (s *Service) HavePermissionTo(ctx context.Context, clubID, userID int64, permission uint64) (bool, error) {
	const op = "service.accessControl.HavePermissionTo"
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

	return domain.HasPermission(userPermissions, permission), nil
}

func (s *Service) CanHandleMembershipRequest(ctx context.Context, clubID, userID int64) (bool, error) {
	return s.HavePermissionTo(ctx, clubID, userID, domain.ManageMembership)
}

func (s *Service) CanManageClub(ctx context.Context, clubID, userID int64) (bool, error) {
	return s.HavePermissionTo(ctx, clubID, userID, domain.ManageClub)
}

func (s *Service) CanManageRoles(ctx context.Context, clubID, userID int64) (bool, error) {
	return s.HavePermissionTo(ctx, clubID, userID, domain.ManageRoles)
}

func (s *Service) CanUpdateRole(ctx context.Context, clubID, userID, roleID int64, permissions []string) (bool, error) {
	const op = "service.accessControl.CanUpdateRole"
	log := s.log.With(slog.String("op", op))

	role, err := s.storage.GetRoleByID(ctx, clubID, roleID)
	if err != nil {
		if errors.Is(err, storage.ErrClubOrRoleNotExists) {
			return false, ErrClubOrRoleNotExists
		}
		log.Error("failed to get role by id", logger.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

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

	if !domain.HasPermission(userPermissions, domain.ManageRoles) {
		return false, fmt.Errorf("%s: %w", domain.Names[domain.ManageRoles], domain.ErrMemberNotHavePermissions)
	}

	userHighestRolePos, err := domain.GetHighestRolePosition(userRoles)
	if err != nil {
		log.Error("failed to get user's highest role position", logger.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if userHighestRolePos <= role.Position {
		return false, fmt.Errorf("%w: %d", domain.ErrMemberNotHavePermissionsToEditRole, role.ID)
	}

	if permissions != nil {
		permissionsHex, err := domain.StringArrToHex(permissions)
		if err != nil {
			log.Error("failed to convert permissions into hexadecimal format", logger.Err(err))
			return false, fmt.Errorf("%s: %w", op, err)
		}

		hasPermissions, err := domain.UserHasPermissions(userPermissions, permissionsHex)
		if err != nil {
			return false, err
		}
		if !hasPermissions {
			return false, nil
		}
	}

	return true, nil

}

func (s *Service) CanEditRoleAndMembersAndDeleteRole(ctx context.Context, clubID, userID, roleID int64) (bool, error) {
	const op = "service.accessControl.canEditRoleAndMembersAndDeleteRole"
	log := s.log.With(slog.String("op", op))

	role, err := s.storage.GetRoleByID(ctx, clubID, roleID)
	if err != nil {
		if errors.Is(err, storage.ErrClubOrRoleNotExists) {
			return false, ErrClubOrRoleNotExists
		}
		log.Error("failed to get role by id", logger.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

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

	if !domain.HasPermission(userPermissions, domain.ManageRoles) {
		return false, fmt.Errorf("%s: %w", domain.Names[domain.ManageRoles], domain.ErrMemberNotHavePermissions)
	}

	userHighestRolePos, err := domain.GetHighestRolePosition(userRoles)
	if err != nil {
		log.Error("failed to get user's highest role position", logger.Err(err))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	if userHighestRolePos <= role.Position {
		return false, fmt.Errorf("%w: %d", domain.ErrMemberNotHavePermissionsToEditRole, role.ID)
	}

	return true, nil
}
