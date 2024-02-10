package main

import (
	"encoding/json"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/config"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/rabbitmq"
	"github.com/rabbitmq/amqp091-go"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting application",
		slog.String("env", cfg.Env),
		slog.Int("port", cfg.GRPC.Port),
	)

	rmq, err := rabbitmq.New(cfg.Rabbitmq, log)
	if err != nil {
		panic(err)
	}

	go func() {
		err = rmq.Consume("club", "user.club.activated", func(msg amqp091.Delivery) error {
			const op = "rabbitmq.user.activated"

			var input struct {
				ID        int64  `json:"id"`
				Email     string `json:"email"`
				FirstName string `json:"first_name"`
				LastName  string `json:"last_name"`
				Barcode   string `json:"barcode"`
				AvatarURL string `json:"avatar_url"`
			}
			err := json.Unmarshal(msg.Body, &input)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			log.Info("input", slog.AnyValue(input))
			return nil
		})
		if err != nil {
			log.Error("failed to consume ")
		}
	}()

	go func() {
		err = rmq.Consume("club", "user.club.updated", func(msg amqp091.Delivery) error {
			const op = "rabbitmq.user.updated"

			var input struct {
				ID        int64   `json:"id"`
				FirstName *string `json:"first_name"`
				LastName  *string `json:"last_name"`
				AvatarURL *string `json:"avatar_url"`
				Major     *string `json:"major"`
				GroupName *string `json:"group_name"`
				Year      *int    `json:"year"`
			}

			err := json.Unmarshal(msg.Body, &input)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			log.Info("input", slog.AnyValue(input))
			return nil
		})
		if err != nil {
			log.Error("failed to consume ")
		}
	}()

	go func() {
		err = rmq.Consume("club", "user.club.deleted", func(msg amqp091.Delivery) error {
			const op = "rabbitmq.user.deleted"

			var input int64

			err := json.Unmarshal(msg.Body, &input)
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}

			log.Info("input", slog.AnyValue(input))
			return nil
		})
		if err != nil {
			log.Error("failed to consume ")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("stopping application", slog.String("signal", sign.String()))
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
