package consumer

// github.com/cespare/xxhash/v2 v2.1.2 // indirect
// 	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
// 	github.com/go-redis/redis/v8 v8.11.5 // indirect
// 	github.com/google/uuid v1.6.0 // indirect
// 	github.com/joho/godotenv v1.5.1 // indirect
// 	github.com/rabbitmq/amqp091-go v1.10.0 // indirect

import (
	"catalog-service/cache"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/google/uuid"
	
)

type FetchProduct struct {
	Id            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	ImageUrl      string    `json:"image_url"`
	Category      string    `json:"category"`
	Description   string    `json:"description"`
	Price         float64   `json:"price"`
	StockQuantity int       `json:"stock_quantity"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func Consumer(repo *cache.RedisClient, conn *amqp.Connection) error {

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
		return fmt.Errorf("Failed to open a channel: %v", err)
	}
	exchangeName := "product_event"
	queueName := "product_queue"
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
		"catalog",
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

	go repo.Listen(context.Background())

	for msg := range msgs {

		var data map[string]interface{}
		err := json.Unmarshal(msg.Body, &data)
		if err != nil {
			fmt.Println("error: ", err)
		}

		event, _ := data["command"].(string)
		payload, _ := data["payload"].(map[string]interface{})
		fmt.Println(data)
		switch event {
		case "product_insert":
			repo.Add <- payload
		case "product_update":
			repo.Update <- payload
		case "product_delete":
			repo.Pop <- payload
		}

	}

	select {}
}
