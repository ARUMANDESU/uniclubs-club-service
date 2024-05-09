package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/config"
	"github.com/ARUMANDESU/uniclubs-club-service/pkg/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"log/slog"
)

const (
	ClubExchangeName             = "club-exchange"
	UserExchangeName             = "user-exchange"
	UserEventsQueue              = "user-events-club-queue"
	PushNotificationRoutingKey   = "user.notification.push"
	UserEventsRoutingKey         = "user.event.*"
	UserUpdatedEventRoutingKey   = "user.event.updated"
	UserActivatedEventRoutingKey = "user.event.activated"
	UserDeletedEventRoutingKey   = "user.event.deleted"
	ClubEventUpdatedRoutingKey   = "club.event.updated"
	ClubEventActivatedRoutingKey = "club.event.activated"
)

type Handler func(msg amqp.Delivery) error

type Rabbitmq struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	cfg  config.Rabbitmq
	log  *slog.Logger
}

func New(cfg config.Rabbitmq, log *slog.Logger) (*Rabbitmq, error) {
	const op = "Rabbitmq.New"

	connString := fmt.Sprintf("amqp://%v:%v@%v:%v/", cfg.User, cfg.Password, cfg.Host, cfg.Port)
	conn, err := amqp.Dial(connString)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to amqp server: %w", op, err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open a channel: %w", op, err)
	}

	err = ch.ExchangeDeclare(
		ClubExchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Error("failed to declare exchange", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = ch.QueueDeclare(
		UserEventsQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Error("failed to declare queue", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = ch.QueueBind(
		UserEventsQueue,
		UserEventsRoutingKey,
		UserExchangeName,
		false,
		nil,
	)
	if err != nil {
		log.Error("failed to bind queue", logger.Err(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Rabbitmq{
		conn: conn,
		ch:   ch,
		cfg:  cfg,
		log:  log,
	}, nil
}

func (r *Rabbitmq) Consume(queue, routingKey string, handler func(msg amqp.Delivery) error) error {
	const op = "Rabbitmq.Consume"
	log := r.log.With(
		slog.String("op", op),
		slog.With("queue", queue),
	)

	err := r.ch.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		log.Error("failed to set Qos", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	msgs, err := r.ch.Consume(
		queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Error("failed to register as consumer", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Debug("routing key", slog.String("key", d.RoutingKey))
			if d.RoutingKey != routingKey {
				err := d.Reject(true)
				if err != nil {
					log.Warn("failed to negatively acknowledge", logger.Err(err))
				}
				continue
			}

			err = handler(d)
			if err != nil {
				log.Warn("failed to handle message", logger.Err(err))
				continue
			}
			err = d.Ack(false)
			if err != nil {
				log.Warn("failed to send an acknowledgement", logger.Err(err))
			}

		}
	}()

	<-forever

	return nil
}

func (r *Rabbitmq) Publish(ctx context.Context, exchangeName string, routingKey string, msg any) error {
	const op = "Rabbitmq.Publish"

	bytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = r.ch.PublishWithContext(
		ctx,
		exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         bytes,
		})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Rabbitmq) Close() error {
	const op = "Rabbitmq.Close"
	log := r.log.With(slog.String("op", op))

	err := r.ch.Close()
	if err != nil {
		log.Error("failed to close channel", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	err = r.conn.Close()
	if err != nil {
		log.Error("failed to close connection", logger.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
