package app

import (
	amqpapp "github.com/ARUMANDESU/uniclubs-club-service/internal/app/amqp"
	grpcapp "github.com/ARUMANDESU/uniclubs-club-service/internal/app/grpc"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/clients/awsS3"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/config"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/rabbitmq"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/services/accessControl"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/services/info"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/services/management"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/services/membership"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/services/user"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/storage/postgresql"
	"github.com/ARUMANDESU/uniclubs-club-service/pkg/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	"log/slog"
)

type App struct {
	GRPCSrv *grpcapp.App
	AMQPApp *amqpapp.App
}

func New(log *slog.Logger, cfg *config.Config, awsCfg aws.Config) *App {
	const op = "app.new"
	l := log.With(slog.String("op", op))

	storage, err := postgresql.New(cfg.DatabaseDSN)
	if err != nil {
		l.Error("failed to connect to postgresql", logger.Err(err))
		panic(err)
	}

	rmq, err := rabbitmq.New(cfg.Rabbitmq, log)
	if err != nil {
		l.Error("failed to connect to rabbitmq", logger.Err(err))
		panic(err)
	}

	imageStorage, err := awsS3.New(awsCfg, cfg.AWS)
	if err != nil {
		l.Error("failed to create aws s3 client", logger.Err(err))
		panic(err)
	}

	usrService := user.New(log, storage)
	managementService := management.New(log, storage, imageStorage, rmq)
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
