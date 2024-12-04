package publishers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	Conn      *amqp.Connection
	Notify    chan map[string]interface{}
	PushOrder chan map[string]interface{}
	Ctx       context.Context
}

func NewProducer(conn *amqp.Connection, ctx context.Context) *Producer {
	return &Producer{
		Conn:      conn,
		Notify:    make(chan map[string]interface{}, 10),
		PushOrder: make(chan map[string]interface{}, 10),
		Ctx:       ctx,
	}
}

func (producer *Producer) Worker() error {

	ticker := time.NewTicker(30 * time.Second)

	defer func() {
		ticker.Stop()
		close(producer.Notify)
	}()

	for {
		select {
		case <-producer.Ctx.Done():
			return producer.Ctx.Err()
		case <-ticker.C:
		case order, ok := <-producer.PushOrder:
			if ok {
				child, cancel := context.WithTimeout(producer.Ctx, 2 * time.Second)
				producer.OrderPublisher(order, child)
				cancel()
			}
		case msg, ok := <-producer.Notify:
			if ok {
				child, cancel := context.WithTimeout(producer.Ctx, 3 * time.Second)
				producer.NotifyPublisher(msg, child)
				cancel()
			}
		}
	}

}


func (producer *Producer) OrderPublisher(payload map[string]interface{}, ctx context.Context) error {
	ch, err := producer.Conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
		return fmt.Errorf("Failed to open a channel: %v", err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	queue, err := ch.QueueDeclare(
		"order_queue", 
		true,          
		false,         
		false,         
		false,         
		nil,           
	)
	if err != nil {
		log.Fatalf("Queue oluşturulamadi: %v", err)
	}

	if err := ch.PublishWithContext(context.Background(),
		"",           
		queue.Name,   
		false,        
		false,        
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	); err != nil {
		return err
	}

	log.Printf("UserDeleted event published: %s", string(body))
	return nil
}

func (producer *Producer) NotifyPublisher(payload map[string]interface{}, ctx context.Context) error {
	ch, err := producer.Conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
		return fmt.Errorf("Failed to open a channel: %v", err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if err := ch.ExchangeDeclare(
		"notifications",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		log.Fatalf("Failed to declare an exchange: %v", err)
	}

	if err := ch.PublishWithContext(context.Background(),
		"notifications",
		"ws",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	); err != nil {
		return err
	}

	log.Printf("UserDeleted event published: %s", string(body))
	return nil
}
