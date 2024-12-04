package publisher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
    amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	Conn *amqp.Connection
}

func (p *Producer) Publisher(message bytes.Buffer, to string) error {

	ch, err := p.Conn.Channel()

	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	q, err := ch.QueueDeclare(
		"mail_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer func() {
		cancel()
		ch.Close()
	}()

	body, err := json.Marshal(map[string]interface{}{
		"template": message.String(),
		"to":       to,
	})

	if err != nil {
		return err
	}

	err = ch.PublishWithContext(ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
