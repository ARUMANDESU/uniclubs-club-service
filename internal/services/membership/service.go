package membership

import (
	"context"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/storage"
	"github.com/ARUMANDESU/uniclubs-club-service/pkg/logger"
	"log/slog"
)

var (
	ErrClubOrRoleNotExists = errors.New("club or role does not exists")
	ErrEditConflict        = errors.New("edit conflict")
)

type Service struct {
	log     *slog.Logger
	storage Storage
}

type Storage interface {
	InsertJoinRequest(ctx context.Context, userID, clubID int64) error
	GetMemberByID(ctx context.Context, clubID, userID int64) (*domain.User, error)
	AddNewMember(ctx context.Context, clubID, userID int64) error
	DeleteJoinRequest(ctx context.Context, clubID, userID int64) error
	CreateRole(ctx context.Context, dto dtos.CreateRoleDTO) (*domain.Role, error)
	DeleteRoleByID(ctx context.Context, clubID, roleID int64) error
	GetRoleByID(ctx context.Context, clubID, roleID int64) (*domain.Role, error)
	UpdateRole(ctx context.Context, role *domain.Role) error
	ChangeRolesPosition(ctx context.Context, clubId int64, dto []*dtos.ChangeRolesPositionDTO) error
	GetRolesOfClubByID(ctx context.Context, clubID int64) ([]*domain.Role, error)
	AddRoleMembers(ctx context.Context, clubID, roleID int64, usersID []int64) error
	RemoveMemberFromClub(ctx context.Context, clubID, userID int64) error
	HaveUserJoinRequest(ctx context.Context, clubID, userID int64) (bool, error)
	RemoveRoleMembers(ctx context.Context, clubID, roleID int64, usersID []int64) error
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

	member, err := s.storage.GetMemberByID(ctx, clubID, userID)
	if err != nil && !errors.Is(err, domain.ErrUserNotClubMember) {
		log.Error("failed to get club member", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	if member != nil {
		return domain.ErrUserAlreadyClubMember
	}

	haveUserJoinRequest, err := s.storage.HaveUserJoinRequest(ctx, clubID, userID)
	if err != nil {
		log.Error("failed to get info about user join request")
		return fmt.Errorf("%s: %w", op, err)
	}
	if haveUserJoinRequest {
		return domain.ErrUserAlreadySentJoinRequest
	}

	err = s.storage.InsertJoinRequest(ctx, userID, clubID)
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

func (s Service) CreateNewRole(ctx context.Context, dto dtos.CreateRoleDTO) (*domain.Role, error) {
	const op = "services.membership.CreateNewRole"
	log := s.log.With(slog.String("op", op))

	role, err := s.storage.CreateRole(ctx, dto)
	if err != nil {
		log.Error("failed to create new role", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return role, nil
}

func (s Service) DeleteRole(ctx context.Context, clubID, roleID int64) error {
	const op = "services.membership.DeleteRole"
	log := s.log.With(slog.String("op", op))

	err := s.storage.DeleteRoleByID(ctx, clubID, roleID)
	if err != nil {
		if errors.Is(err, storage.ErrClubOrRoleNotExists) {
			return fmt.Errorf("%s: %w", op, ErrClubOrRoleNotExists)
		}
		log.Error("failed to delete a role", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s Service) GetRole(ctx context.Context, clubID, roleID int64) (*domain.Role, error) {
	const op = "services.membership.GetRole"
	log := s.log.With(slog.String("op", op))

	role, err := s.storage.GetRoleByID(ctx, clubID, roleID)
	if err != nil {
		if errors.Is(err, storage.ErrClubOrRoleNotExists) {
			return nil, fmt.Errorf("%s: %w", op, ErrClubOrRoleNotExists)
		}
		log.Error("failed to get the role", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return role, nil
}

func (s Service) UpdateRole(ctx context.Context, role *domain.Role) error {
	const op = "services.membership.UpdateRole"
	log := s.log.With(slog.String("op", op))

	err := s.storage.UpdateRole(ctx, role)
	if err != nil {
		if errors.Is(err, storage.ErrEditConflict) {
			return fmt.Errorf("%s: %w", op, ErrEditConflict)
		}
		log.Error("failed to update the role", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s Service) ChangeRolesPosition(ctx context.Context, clubID int64, dto []*dtos.ChangeRolesPositionDTO) ([]*domain.Role, error) {
	const op = "services.membership.ChangeRolesPosition"
	log := s.log.With(slog.String("op", op))

	err := s.storage.ChangeRolesPosition(ctx, clubID, dto)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCannotEditRoleMember), errors.Is(err, domain.ErrClubOrRoleNotExists):
			return nil, err
		default:
			log.Error("failed to change roles positions", logger.Err(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}

	}

	roles, err := s.storage.GetRolesOfClubByID(ctx, clubID)
	if err != nil {
		if errors.Is(err, storage.ErrClubOrRoleNotExists) {
			return nil, fmt.Errorf("%s: %w", op, ErrClubOrRoleNotExists)
		}
		log.Error("failed to get roles of the club", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return roles, nil
}

func (s Service) AddRoleMembers(ctx context.Context, clubID, roleID int64, usersID []int64) error {
	const op = "services.membership.AddRoleMembers"
	log := s.log.With(slog.String("op", op))

	err := s.storage.AddRoleMembers(ctx, clubID, roleID, usersID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserAlreadyRoleMember):
			return err
		case errors.Is(err, storage.ErrUserNotClubMember):
			return err
		default:
			log.Error("failed to add new role members", logger.Err(err))
			return fmt.Errorf("%s: %w", op, err)
		}

	}

	return nil

}

func (s Service) RemoveMemberFromClub(ctx context.Context, clubID, userID int64) error {
	const op = "services.membership.RemoveMemberFromClub"
	log := s.log.With(slog.String("op", op))

	err := s.storage.RemoveMemberFromClub(ctx, clubID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrMemberNotFound) {
			log.Debug("err member not found", logger.Err(err))
			return err
		}
		log.Error("failed to remove member from club", logger.Err(err))
		return err
	}

	return nil
}

func (s Service) RemoveRoleMembers(ctx context.Context, clubID, roleID int64, usersID []int64) error {
	const op = "services.membership.RemoveRoleMembers"
	log := s.log.With(slog.String("op", op))

	err := s.storage.RemoveRoleMembers(ctx, clubID, roleID, usersID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotClubMember),
			errors.Is(err, domain.ErrCannotEditRoleMember):
			return err
		default:
			log.Error("failed to add new role members", logger.Err(err))
			return fmt.Errorf("%s: %w", op, err)
		}

	}

	return nil
}
