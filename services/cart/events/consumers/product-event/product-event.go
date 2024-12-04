package product_event

import (
	"cart-service/service"
	"context"
	"encoding/json"
	"fmt"
	"log"
	amqp "github.com/rabbitmq/amqp091-go"
)

func Consume(repo *service.RedisClient, conn *amqp.Connection) error {

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
		return fmt.Errorf("Failed to open a channel: %v", err)
	}
	exchangeName := "product_event"
	queueName := "cart_queue"
	defer func() {
		ch.Close()
		conn.Close()
	}()

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

	go repo.EventListener(context.Background())

	for msg := range msgs {

		var data map[string]interface{}
		err := json.Unmarshal(msg.Body, &data)
		if err != nil {
			fmt.Printf("error:%s", err)
		}
		
		repo.ProductEvents <- data
	
	}

	select {}
}
