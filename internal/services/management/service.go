package management

import (
	"context"
	"errors"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/clients/image"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/domain/dtos"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/storage"
	"github.com/ARUMANDESU/uniclubs-club-service/pkg/logger"
	imagev1 "github.com/ARUMANDESU/uniclubs-protos/gen/go/filestorage"
	"log/slog"
)

var (
	ErrFailedToBeginTx = errors.New("failed to begin transaction")
	ErrClubNotExists   = errors.New("club does not exists")
	ErrEditConflict    = errors.New("edit conflict")
)

type Service struct {
	log         *slog.Logger
	storage     Storage
	imageClient *image.Client
}

type Storage interface {
	SaveClub(ctx context.Context, dto dtos.CreateClubDTO) error
	ApproveClub(ctx context.Context, clubID int64) error
	RejectClub(ctx context.Context, clubID int64) error
	UpdateClub(ctx context.Context, club *domain.Club) error
	GetClubByID(ctx context.Context, clubID int64) (*domain.Club, error)
}

func New(log *slog.Logger, storage Storage, imageClient *image.Client) *Service {
	return &Service{
		log:         log,
		storage:     storage,
		imageClient: imageClient,
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

func (s Service) UpdateLogo(ctx context.Context, clubID int64, logo []byte) (*domain.Club, error) {
	const op = "services.management.UpdateLogo"
	log := s.log.With(slog.String("op", op))

	req, err := s.imageClient.UploadImage(ctx, &imagev1.UploadImageRequest{Image: logo, Filename: fmt.Sprintf("club-%d-logo", clubID)})
	if err != nil {
		log.Error("failed to update logo", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	club, err := s.storage.GetClubByID(ctx, clubID)
	if err != nil {
		if errors.Is(err, storage.ErrClubNotExists) {
			log.Error("club does not exists", logger.Err(err))
			return nil, fmt.Errorf("%s: %w", op, ErrClubNotExists)
		}
		log.Error("failed to get club by ID", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	club.LogoURL = req.ImageUrl

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
