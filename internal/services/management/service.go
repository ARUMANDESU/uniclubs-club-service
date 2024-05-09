package management

import (
	"context"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/rabbitmq"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/storage"
	imageUtils "github.com/ARUMANDESU/uniclubs-club-service/pkg/image"
	"github.com/ARUMANDESU/uniclubs-club-service/pkg/logger"
	"log/slog"
	"path"
	"time"
)

var (
	ErrClubNotExists = errors.New("club does not exists")
	ErrEditConflict  = errors.New("edit conflict")
)

type Service struct {
	log          *slog.Logger
	amqp         Amqp
	storage      Storage
	imageStorage ImageStorage
}

type Amqp interface {
	Publish(ctx context.Context, exchangeName string, routingKey string, msg any) error
}

type ImageStorage interface {
	UploadImage(ctx context.Context, image []byte, filename string) (string, error)
	DeleteImage(ctx context.Context, filename string) error
}

type Storage interface {
	SaveClub(ctx context.Context, dto dtos.CreateClubDTO) error
	ApproveClub(ctx context.Context, clubID int64) (int64, error)
	RejectClub(ctx context.Context, clubID int64) error
	UpdateClub(ctx context.Context, club *domain.Club) error
	GetClubByID(ctx context.Context, clubID int64, isApproved bool) (*domain.Club, error)
	GetUserRoles(ctx context.Context, clubID, userID int64) (roles []*domain.Role, isOwner bool, err error)
}

func New(log *slog.Logger, storage Storage, imageStorage ImageStorage, amqp Amqp) *Service {
	return &Service{
		log:          log,
		amqp:         amqp,
		storage:      storage,
		imageStorage: imageStorage,
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
			"clubID": club.ID,
			"name":   club.Name,
			"logo":   club.LogoURL,
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

func (s Service) UpdateLogo(ctx context.Context, clubID int64, logo []byte) (*domain.Club, error) {
	const op = "services.management.UpdateLogo"
	log := s.log.With(slog.String("op", op))

	club, err := s.storage.GetClubByID(ctx, clubID, true)
	if err != nil {
		if errors.Is(err, storage.ErrClubNotExists) {
			log.Error("club does not exists", logger.Err(err))
			return nil, fmt.Errorf("%s: %w", op, ErrClubNotExists)
		}
		log.Error("failed to get club by ID", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Delete previous club logo
	if club.LogoURL != "" {
		// path.Base returns the last element of the path: object key
		err = s.imageStorage.DeleteImage(ctx, path.Base(club.LogoURL))
		if err != nil {
			log.Error("failed to delete previous avatar", logger.Err(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	// Compress image
	compressImage, filename, err := imageUtils.CompressImage(logo, 75)
	if err != nil {
		log.Error("failed to compress image", logger.Err(err))
		return nil, err
	}

	imageCtx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	url, err := s.imageStorage.UploadImage(imageCtx, compressImage, filename)
	if err != nil {
		log.Error("failed to upload avatar", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	club.LogoURL = url

	err = s.storage.UpdateClub(ctx, club)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrEditConflict):
			log.Error("edit club conflict", logger.Err(err))
			return nil, fmt.Errorf("%s: %w", op, ErrEditConflict)
		default:
			log.Error("failed to get club by ID", logger.Err(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	go func() {
		msg := map[string]interface{}{
			"clubID": club.ID,
			"logo":   club.LogoURL,
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

	return club, nil
}

func (s Service) UpdateBanner(ctx context.Context, clubID int64, banner []byte) (*domain.Club, error) {
	const op = "services.management.UpdateBanner"
	log := s.log.With(slog.String("op", op))

	club, err := s.storage.GetClubByID(ctx, clubID, true)
	if err != nil {
		if errors.Is(err, storage.ErrClubNotExists) {
			log.Error("club does not exists", logger.Err(err))
			return nil, fmt.Errorf("%s: %w", op, ErrClubNotExists)
		}
		log.Error("failed to get club by ID", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Delete previous club banner
	if club.LogoURL != "" {
		// path.Base returns the last element of the path: object key
		err = s.imageStorage.DeleteImage(ctx, path.Base(club.LogoURL))
		if err != nil {
			log.Error("failed to delete previous avatar", logger.Err(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	// Compress image
	compressImage, filename, err := imageUtils.CompressImage(banner, 75)
	if err != nil {
		log.Error("failed to compress image", logger.Err(err))
		return nil, err
	}

	imageCtx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	url, err := s.imageStorage.UploadImage(imageCtx, compressImage, filename)
	if err != nil {
		log.Error("failed to upload avatar", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	club.BannerURL = url

	err = s.storage.UpdateClub(ctx, club)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrEditConflict):
			log.Error("edit club conflict", logger.Err(err))
			return nil, fmt.Errorf("%s: %w", op, ErrEditConflict)
		default:
			log.Error("failed to get club by ID", logger.Err(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}

	return club, nil
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
			"clubID": club.ID,
			"name":   club.Name,
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
		return fmt.Errorf("%w: %d", op, domain.ErrUserAlreadyClubOwner, targetID)
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

	go s.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.PushNotificationRoutingKey, notification)

	s.amqp.Publish(ctx, rabbitmq.UserExchangeName, rabbitmq.PushNotificationRoutingKey, notificationToOldOwner)

	return nil
}
