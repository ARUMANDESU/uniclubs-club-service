package app

import (
	"context"
	amqpapp "github.com/ARUMANDESU/uniclubs-club-service/internal/app/amqp"
	grpcapp "github.com/ARUMANDESU/uniclubs-club-service/internal/app/grpc"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/clients/image"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/config"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/rabbitmq"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/services/accessControl"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/services/info"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/services/management"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/services/membership"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/services/user"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/storage/postgresql"
	"github.com/ARUMANDESU/uniclubs-club-service/pkg/logger"
	"log/slog"
)

type App struct {
	GRPCSrv *grpcapp.App
	AMQPApp *amqpapp.App
}

func New(log *slog.Logger, cfg *config.Config) *App {
	const op = "App.New"
	l := log.With(slog.String("op", op))

	storage, err := postgresql.New(cfg.DatabaseDSN)
	if err != nil {
		return nil
	}

	rmq, err := rabbitmq.New(cfg.Rabbitmq, log)
	if err != nil {

	}

	imageClient, err := image.New(context.Background(), log, cfg.Clients)
	if err != nil {
		l.Error("failed to connect to imagestorage service", logger.Err(err))
		panic(err)
	}

	usrService := user.New(log, storage)
	managementService := management.New(log, storage, imageClient, rmq)
	membershipService := membership.New(log, storage, rmq)
	infoService := info.New(log, storage)
	permissionService := accessControl.New(log, storage)

	grpcApp := grpcapp.New(
		log,
		cfg.GRPC.Port,
		managementService,
		membershipService,
		infoService,
		permissionService,
	)
	amqpApp := amqpapp.New(log, usrService, rmq)

	return &App{GRPCSrv: grpcApp, AMQPApp: amqpApp}
}
