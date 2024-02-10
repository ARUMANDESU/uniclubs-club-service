package amqpapp

import (
	"github.com/ARUMANDESU/uniclubs-club-service/internal/rabbitmq"
	"github.com/ARUMANDESU/uniclubs-club-service/pkg/logger"
	"github.com/rabbitmq/amqp091-go"
	"log/slog"
)

type App struct {
	log        *slog.Logger
	amqp       Amqp
	usrService UserService
}

type Amqp interface {
	Consume(queue string, routingKey string, handler func(msg amqp091.Delivery) error) error
}

type UserService interface {
	HandleCreateUser(msg amqp091.Delivery) error
	HandleUpdateUser(msg amqp091.Delivery) error
	HandleDeleteUser(msg amqp091.Delivery) error
}

func New(log *slog.Logger, service UserService, amqp Amqp) *App {
	return &App{
		log:        log,
		amqp:       amqp,
		usrService: service,
	}
}

func (a *App) SetupMessageConsumers() {
	a.consumeMessages("club", "user.club.activated", a.usrService.HandleCreateUser)
	a.consumeMessages("club", "user.club.updated", a.usrService.HandleUpdateUser)
	a.consumeMessages("club", "user.club.deleted", a.usrService.HandleDeleteUser)
}

func (a *App) consumeMessages(queue, routingKey string, handler rabbitmq.Handler) {
	go func() {
		const op = "amqp.app.consumeMessages"
		log := a.log.With(slog.String("op", op))

		err := a.amqp.Consume(queue, routingKey, handler)
		if err != nil {
			log.Error("failed to consume ", logger.Err(err))
		}
	}()
}
