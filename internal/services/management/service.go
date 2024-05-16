package management

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/rabbitmq"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/storage"
	"github.com/ARUMANDESU/uniclubs-club-service/pkg/logger"
)

var (
	ErrClubNotExists = errors.New("club does not exists")
	ErrEditConflict  = errors.New("edit conflict")
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
	SaveClub(ctx context.Context, dto dtos.CreateClubDTO) error
	ApproveClub(ctx context.Context, clubID int64) (int64, error)
	RejectClub(ctx context.Context, clubID int64) error
	UpdateClub(ctx context.Context, club *domain.Club) error
	GetClubByID(ctx context.Context, clubID int64, isApproved bool) (*domain.Club, error)
	GetUserRoles(ctx context.Context, clubID, userID int64) (roles []*domain.Role, isOwner bool, err error)
	DeleteClubByID(ctx context.Context, clubID int64) error
}

func New(log *slog.Logger, storage Storage, amqp Amqp) *Service {
	return &Service{
		log:     log,
		amqp:    amqp,
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

	club, err := s.storage.GetClubByID(ctx, clubID, false)
	if err != nil {
		if errors.Is(err, storage.ErrClubNotExists) {
			return fmt.Errorf("%s: %w", op, ErrClubNotExists)
		}
		log.Error("failed to get club by ID", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	userID, err := s.storage.ApproveClub(ctx, club.ID)
	if err != nil {
		log.Error("failed to approve club", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	go func() {
		err := s.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.PushNotificationRoutingKey, domain.Notification{
			UserID:      userID,
			Message:     "Your club has been approved",
			Description: "Your club has been approved and is now visible to other users",
			Status:      "NEW",
			Severity:    "INFO",
			Source:      "club",
			DisplayType: "INBOX",
			CreatedAt:   time.Now().String(),
			ExpiryAt:    time.Now().Add(time.Hour * 24 * 2).String(),
		})

		if err != nil {
			log.Error("failed to publish notification", logger.Err(err))
		}

		msg := map[string]interface{}{
			"id":       club.ID,
			"name":     club.Name,
			"logo_url": club.LogoURL,
		}

		err = s.amqp.Publish(
			ctx,
			rabbitmq.ClubExchangeName,
			rabbitmq.ClubEventActivatedRoutingKey,
			msg,
		)
		if err != nil {
			log.Error("failed to publish notification", logger.Err(err))
		}
	}()

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

func (s Service) UpdateLogo(ctx context.Context, clubID int64, logoUrl string) (club *domain.Club, prevLogoUrl string, err error) {
	const op = "services.management.UpdateLogo"
	log := s.log.With(slog.String("op", op))

	club, err = s.storage.GetClubByID(ctx, clubID, true)
	if err != nil {
		if errors.Is(err, storage.ErrClubNotExists) {
			log.Error("club does not exists", logger.Err(err))
			return nil, "", fmt.Errorf("%s: %w", op, ErrClubNotExists)
		}
		log.Error("failed to get club by ID", logger.Err(err))
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	if club.LogoURL != "" {
		prevLogoUrl = club.LogoURL
	}

	club.LogoURL = logoUrl

	err = s.storage.UpdateClub(ctx, club)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrEditConflict):
			log.Error("edit club conflict", logger.Err(err))
			return nil, "", fmt.Errorf("%s: %w", op, ErrEditConflict)
		default:
			log.Error("failed to get club by ID", logger.Err(err))
			return nil, "", fmt.Errorf("%s: %w", op, err)
		}
	}

	go func() {
		msg := map[string]interface{}{
			"id":       club.ID,
			"logo_url": club.LogoURL,
		}

		err = s.amqp.Publish(
			ctx,
			rabbitmq.ClubExchangeName,
			rabbitmq.ClubEventUpdatedRoutingKey,
			msg,
		)
		if err != nil {
			log.Error("failed to publish notification", logger.Err(err))
		}
	}()

	return club, prevLogoUrl, nil
}

func (s Service) UpdateBanner(ctx context.Context, clubID int64, bannerUrl string) (club *domain.Club, prevBannerUrl string, err error) {
	const op = "services.management.UpdateBanner"
	log := s.log.With(slog.String("op", op))

	club, err = s.storage.GetClubByID(ctx, clubID, true)
	if err != nil {
		if errors.Is(err, storage.ErrClubNotExists) {
			log.Error("club does not exists", logger.Err(err))
			return nil, "", fmt.Errorf("%s: %w", op, ErrClubNotExists)
		}
		log.Error("failed to get club by ID", logger.Err(err))
		return nil, "", fmt.Errorf("%s: %w", op, err)
	}

	if club.BannerURL != "" {
		prevBannerUrl = club.BannerURL
	}

	club.BannerURL = bannerUrl

	err = s.storage.UpdateClub(ctx, club)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrEditConflict):
			log.Error("edit club conflict", logger.Err(err))
			return nil, "", fmt.Errorf("%s: %w", op, ErrEditConflict)
		default:
			log.Error("failed to get club by ID", logger.Err(err))
			return nil, "", fmt.Errorf("%s: %w", op, err)
		}
	}

	return club, prevBannerUrl, nil
}

func (s Service) UpdateClub(ctx context.Context, club *domain.Club) error {
	const op = "services.management.UpdateClub"
	log := s.log.With(slog.String("op", op))

	err := s.storage.UpdateClub(ctx, club)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrEditConflict):
			log.Error("edit club conflict", logger.Err(err))
			return fmt.Errorf("%s: %w", op, ErrEditConflict)
		default:
			log.Error("failed to get club by ID", logger.Err(err))
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	go func() {
		msg := map[string]interface{}{
			"id":   club.ID,
			"name": club.Name,
		}

		err = s.amqp.Publish(
			ctx,
			rabbitmq.ClubExchangeName,
			rabbitmq.ClubEventUpdatedRoutingKey,
			msg,
		)
		if err != nil {
			log.Error("failed to publish notification", logger.Err(err))
		}
	}()

	return nil
}

func (s Service) TransferOwnership(ctx context.Context, clubID, userID, targetID int64) error {
	const op = "services.management.TransferOwnership"
	log := s.log.With(slog.String("op", op))

	club, err := s.storage.GetClubByID(ctx, clubID, true)
	if err != nil {
		if errors.Is(err, storage.ErrClubNotExists) {
			log.Error("club does not exists", logger.Err(err))
			return fmt.Errorf("%s: %w", op, ErrClubNotExists)
		}
		log.Error("failed to get club by ID", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	if club.OwnerID != userID {
		return fmt.Errorf("%s: %w", op, domain.ErrUserNotClubOwner)
	}

	_, isOwner, err := s.storage.GetUserRoles(ctx, clubID, targetID)
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
		return fmt.Errorf("%w: %d", domain.ErrUserAlreadyClubOwner, targetID)
	}

	club.OwnerID = targetID

	err = s.storage.UpdateClub(ctx, club)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrEditConflict):
			log.Error("edit club conflict", logger.Err(err))
			return fmt.Errorf("%s: %w", op, ErrEditConflict)
		default:
			log.Error("failed to get club by ID", logger.Err(err))
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	go func() {
		notification := domain.Notification{
			UserID:      targetID,
			Message:     fmt.Sprintf("Now you are the owner of club '%s'", club.Name),
			Description: fmt.Sprintf("You are now the owner of club '%s'", club.Name),
			Status:      "NEW",
			Severity:    "INFO",
			Source:      "club",
			DisplayType: "INBOX",
			CreatedAt:   time.Now().String(),
			ExpiryAt:    time.Now().Add(time.Hour * 24 * 2).String(),
		}
		notificationToOldOwner := domain.Notification{
			UserID:      userID,
			Message:     fmt.Sprintf("You are no longer the owner of club '%s'", club.Name),
			Description: fmt.Sprintf("You successfully transferred ownership of club '%s' to another user", club.Name),
			Status:      "NEW",
			Severity:    "INFO",
			Source:      "club",
			DisplayType: "INBOX",
			CreatedAt:   time.Now().String(),
			ExpiryAt:    time.Now().Add(time.Hour * 24 * 2).String(),
		}
		err := s.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.PushNotificationRoutingKey, notification)
		if err != nil {
			log.Warn("failed to send notification", logger.Err(err))
		}
		err = s.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.PushNotificationRoutingKey, notificationToOldOwner)
		if err != nil {
			log.Warn("failed to send notification", logger.Err(err))
		}
	}()

	return nil
}

func (s Service) DeleteClub(ctx context.Context, clubID int64) error {
	const op = "services.management.DeleteClubByID"
	log := s.log.With(slog.String("op", op))

	err := s.storage.DeleteClubByID(ctx, clubID)

	if err != nil {
		switch {
		case errors.Is(err, storage.ErrClubNotExists):
			log.Error("club does not exists", logger.Err(err))
			return fmt.Errorf("%s: %w", op, ErrClubNotExists)
		case errors.Is(err, domain.ErrUserNotClubOwner):
			log.Error("user is not club owner", logger.Err(err))
			return fmt.Errorf("%s: %w", op, domain.ErrUserNotClubOwner)
		default:
			log.Error("failed to delete club", logger.Err(err))
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}
