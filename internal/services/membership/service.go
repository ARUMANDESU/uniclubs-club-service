package membership

import (
	"context"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/rabbitmq"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/services/info"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/storage"
	"github.com/ARUMANDESU/uniclubs-club-service/pkg/logger"
	"log/slog"
	"time"
)

var (
	ErrClubOrRoleNotExists = errors.New("club or role does not exists")
	ErrEditConflict        = errors.New("edit conflict")
)

type Service struct {
	log     *slog.Logger
	amqp    Amqp
	storage Storage
}

type Amqp interface {
	Publish(ctx context.Context, exchangeName string, routingKey string, msg any) error
}

type Storage interface {
	GetClubByID(ctx context.Context, clubID int64, isApproved bool) (*domain.Club, error)
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
	GetUserRoles(ctx context.Context, clubID, userID int64) (roles []*domain.Role, isOwner bool, err error)
	BanMember(ctx context.Context, dto dtos.BanMemberDTO) error
	UnbanUser(ctx context.Context, dto dtos.UnbanUserDTO) error
	GetBanRecord(ctx context.Context, clubID, userID int64) (*domain.BanRecord, error)
}

func New(log *slog.Logger, storage Storage, amqp Amqp) *Service {
	return &Service{
		log:     log,
		storage: storage,
		amqp:    amqp,
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

	banRecord, err := s.storage.GetBanRecord(ctx, clubID, userID)
	if err != nil && !errors.Is(err, storage.ErrBanRecordNotExists) {
		log.Error("failed to get ban record", logger.Err(err))
		return err
	}
	if banRecord != nil {
		return fmt.Errorf("cannot join club because %w", domain.ErrUserBanned)
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

	_, isOwner, err := s.storage.GetUserRoles(ctx, clubID, userID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotClubMember):
			return fmt.Errorf("%s: %w", op, domain.ErrUserNotClubMember)
		default:
			log.Error("failed to get user roles", logger.Err(err))
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if isOwner {
		return domain.ErrOwnerCannotLeaveClub
	}

	err = s.storage.RemoveMemberFromClub(ctx, clubID, userID)
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

func (s Service) BanMember(ctx context.Context, dto dtos.BanMemberDTO) error {
	const op = "services.membership.BanMember"
	log := s.log.With(slog.String("op", op))

	_, isOwner, err := s.storage.GetUserRoles(ctx, dto.ClubID, dto.UserID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotClubMember):
			return fmt.Errorf("%s: %w", op, domain.ErrUserNotClubMember)
		default:
			log.Error("failed to get user roles", logger.Err(err))
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if isOwner {
		return domain.ErrOwnerCannotBeBanned
	}

	err = s.storage.RemoveMemberFromClub(ctx, dto.ClubID, dto.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrMemberNotFound) {
			log.Debug("err member not found", logger.Err(err))
			return err
		}
		log.Error("failed to remove member from club", logger.Err(err))
		return err
	}

	err = s.storage.BanMember(ctx, dto)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserAlreadyBanned):
			return domain.ErrUserAlreadyBanned
		default:
			log.Error("failed to ban member", logger.Err(err))
			return err
		}
	}

	club, err := s.storage.GetClubByID(ctx, dto.ClubID, true)
	if err != nil {
		if errors.Is(err, storage.ErrClubNotExists) {
			return fmt.Errorf("%s: %w", op, info.ErrClubNotExists)
		}
		log.Error("failed to get club by ID", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	notification := domain.Notification{
		UserID:      dto.UserID,
		Message:     fmt.Sprintf("You have been banned from the club %s", club.Name),
		Description: fmt.Sprintf("You have been banned from the club %s, because %s", club.Name, dto.Reason),
		Status:      "NEW",
		Severity:    "HIGH",
		Source:      "club",
		DisplayType: "INBOX",
		CreatedAt:   time.Now().String(),
		ExpiryAt:    time.Now().Add(time.Hour * 24 * 2).String(),
	}

	err = s.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.PushNotificationRoutingKey, notification)
	if err != nil {
		log.Error("failed to publish notification", logger.Err(err))
		return err
	}

	return nil
}

func (s Service) GetBanRecord(ctx context.Context, clubID, userID int64) (*domain.BanRecord, error) {
	const op = "services.membership.GetBanRecord"
	log := s.log.With(slog.String("op", op))

	banRecord, err := s.storage.GetBanRecord(ctx, clubID, userID)
	if err != nil {
		if errors.Is(err, storage.ErrBanRecordNotExists) {
			return nil, domain.ErrUserNotBanned
		}
		log.Error("failed to get ban record", logger.Err(err))
		return nil, err
	}

	return banRecord, nil
}

func (s Service) UnbanUser(ctx context.Context, dto dtos.UnbanUserDTO) error {
	const op = "services.membership.UnbanUser"
	log := s.log.With(slog.String("op", op))

	err := s.storage.UnbanUser(ctx, dto)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotBanned):
			return domain.ErrUserNotBanned
		default:
			log.Error("failed to unban user", logger.Err(err))
			return err
		}
	}

	club, err := s.storage.GetClubByID(ctx, dto.ClubID, true)
	if err != nil {
		if errors.Is(err, storage.ErrClubNotExists) {
			return fmt.Errorf("%s: %w", op, info.ErrClubNotExists)
		}
		log.Error("failed to get club by ID", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	notification := domain.Notification{
		UserID:      dto.UserID,
		Message:     fmt.Sprintf("You have been unbanned from the club %s", club.Name),
		Description: fmt.Sprintf("You have been unbanned from the club %s", club.Name),
		Status:      "NEW",
		Severity:    "HIGH",
		Source:      "club",
		DisplayType: "INBOX",
		CreatedAt:   time.Now().String(),
		ExpiryAt:    time.Now().Add(time.Hour * 24 * 2).String(),
	}

	err = s.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.PushNotificationRoutingKey, notification)
	if err != nil {
		log.Error("failed to publish notification", logger.Err(err))
		return err
	}

	return nil
}
