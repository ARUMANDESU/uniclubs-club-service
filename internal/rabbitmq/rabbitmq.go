package rabbitmq

import (
	"fmt"
	"github.com/ARUMANDESU/uniclubs-club-service/internal/config"
	"github.com/ARUMANDESU/uniclubs-club-service/pkg/logger"
	"github.com/rabbitmq/amqp091-go"
	"log/slog"
)

type Handler func(msg amqp091.Delivery) error

type Rabbitmq struct {
	conn *amqp091.Connection
	ch   *amqp091.Channel
	cfg  config.Rabbitmq
	log  *slog.Logger
}

func New(cfg config.Rabbitmq, log *slog.Logger) (*Rabbitmq, error) {
	const op = "Rabbitmq.New"

	connString := fmt.Sprintf("amqp://%v:%v@%v:%v/", cfg.User, cfg.Password, cfg.Host, cfg.Port)
	conn, err := amqp091.Dial(connString)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to amqp server: %w", op, err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open a channel: %w", op, err)
	}

	return &Rabbitmq{
		conn: conn,
		ch:   ch,
		cfg:  cfg,
		log:  log,
	}, nil
}

func (r *Rabbitmq) Consume(queue, routingKey string, handler Handler) error {
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
