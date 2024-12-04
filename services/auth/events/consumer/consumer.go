package consumer

import (
	"auth-service/service"
	"encoding/json"
	"fmt"
	"log"
	amqp "github.com/rabbitmq/amqp091-go"
	"context"
)

func Consume(s service.IAuthService, conn *amqp.Connection) error {

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
		return fmt.Errorf("Failed to open a channel: %v", err)
	}

	exchangeName := "user-delete-event"
	queueName := "auth_user_event"

	defer ch.Close()
	

	if err := ch.ExchangeDeclare(
		exchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %v", err)
	}

	err = ch.QueueBind(
		q.Name,
		"",
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind a queue: %v", err)
		return fmt.Errorf("failed to bind queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %v", err)
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for msg := range msgs {
			var data map[string]interface{}
			err := json.Unmarshal(msg.Body, &data)
			if err != nil {
				log.Printf("error:%s", err)
			}
			id, ok := data["id"].(string)
			if ok {
				if err := s.DeleteUserById(ctx, id); err != nil {
					log.Printf("uuid parse error:%s", err)
				}
			}
		}
	}()

	<-ctx.Done()
	return nil
}
