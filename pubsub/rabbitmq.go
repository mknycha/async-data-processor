package pubsub

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

const baseQueueName = "task_queue"

type pubsubWrapper struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewWrapper(conn *amqp.Connection) (*pubsubWrapper, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "failed to open a channel")
	}
	return &pubsubWrapper{
		conn: conn,
		ch:   ch,
	}, nil
}

func (w *pubsubWrapper) PublishWithContext(ctx context.Context, messageBody []byte, shardNo int) error {
	err := w.ch.PublishWithContext(ctx,
		"",
		queueName(shardNo),
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         messageBody,
		})
	if err != nil {
		return errors.Wrap(err, "failed to publish a message")
	}
	return nil
}

func (w *pubsubWrapper) MessagesChannel(shardNo int) (<-chan amqp.Delivery, error) {
	err := w.ch.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set QoS")
	}

	msgs, err := w.ch.Consume(
		queueName(shardNo),
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to register a consumer")
	}

	return msgs, nil
}

func (w *pubsubWrapper) QueueDeclare(shardNo int) error {
	_, err := w.ch.QueueDeclare(
		queueName(shardNo),
		// durable is set to true, so that messages are not lost if connection to channel is lost
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "failed to declare a queue")
	}
	return err
}

func queueName(shardNo int) string {
	return fmt.Sprintf("%s_%d", baseQueueName, shardNo)
}
