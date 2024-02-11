package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/storage"
	"github.com/ARUMANDESU/uniclubs-club-service/pkg/logger"
	"github.com/rabbitmq/amqp091-go"
	"log/slog"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotExist = errors.New("user does not exist")
)

type Service struct {
	log        *slog.Logger
	usrStorage Storage
}

type Storage interface {
	SaveUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, userID int64) (user *domain.User, err error)
	UpdateUser(ctx context.Context, user *domain.User) error
	DeleteUserByID(ctx context.Context, userID int64) error
}

func New(log *slog.Logger, storage Storage) *Service {
	return &Service{
		log:        log,
		usrStorage: storage,
	}
}

func (s Service) HandleCreateUser(msg amqp091.Delivery) error {
	const op = "rabbitmq.user.activated"

	log := s.log.With(slog.String("op", op))

	var input domain.User
	err := json.Unmarshal(msg.Body, &input)
	if err != nil {
		log.Error("failed to unmarshal message", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	err = s.usrStorage.SaveUser(context.Background(), &input)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserExists):
			log.Error("user already exists", logger.Err(err))
			return fmt.Errorf("%s: %w", op, ErrUserExists)

		default:
			log.Error("failed to save user", logger.Err(err))
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil

}

func (s Service) HandleUpdateUser(msg amqp091.Delivery) error {
	const op = "rabbitmq.user.activated"

	log := s.log.With(slog.String("op", op))

	var input struct {
		ID        int64   `json:"id"`
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
		AvatarURL *string `json:"avatar_url"`
	}

	err := json.Unmarshal(msg.Body, &input)
	if err != nil {
		log.Error("failed to unmarshal message", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user, err := s.usrStorage.GetUserByID(ctx, input.ID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			log.Error("user does not exists", logger.Err(err))
			return fmt.Errorf("%s: %w", op, ErrUserNotExist)
		default:
			log.Error("failed to get user", logger.Err(err))
			return fmt.Errorf("%s: %w", op, err)
		}

	}

	if input.FirstName != nil {
		user.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		user.LastName = *input.LastName
	}
	if input.AvatarURL != nil {
		user.AvatarURL = *input.AvatarURL
	}

	err = s.usrStorage.UpdateUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			log.Error("user not found", logger.Err(err))
			return fmt.Errorf("%s: %w", op, ErrUserNotExist)
		default:
			log.Error("failed to update user", logger.Err(err))
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (s Service) HandleDeleteUser(msg amqp091.Delivery) error {
	const op = "rabbitmq.user.activated"

	log := s.log.With(slog.String("op", op))

	var userID int64

	err := json.Unmarshal(msg.Body, &userID)
	if err != nil {
		log.Error("failed to unmarshal message", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	err = s.usrStorage.DeleteUserByID(context.Background(), userID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrUserNotExists):
			log.Error("user not found", logger.Err(err))
			return fmt.Errorf("%s: %w", op, ErrUserNotExist)
		default:
			log.Error("failed to delete user", logger.Err(err))
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}
