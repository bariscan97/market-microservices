package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"order-service/service"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Consume(conn *amqp.Connection, repo *service.OrderService) error {

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Kanal oluşturulamadi: %v", err)
	}
	defer channel.Close()

	queue, err := channel.QueueDeclare(
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

	messages, err := channel.Consume(
		queue.Name, 
		"",         
		true,       
		false,      
		false,      
		false,      
		nil,       
	)
	if err != nil {
		log.Fatalf("Mesaj alinamadi: %v", err)
	}

	fmt.Println("Mesajlar dinleniyor...")

	ctx := context.Background()

	for {
		select {
		case <- ctx.Done():
			fmt.Println("ctx Done")
			return ctx.Err()
		case msg := <-messages:
			
			var data map[string]interface{}
			if err := json.Unmarshal(msg.Body, &data); err != nil {
				log.Printf("err: %s", err)
			}
			fmt.Println("msg:",data)
			child, cancel := context.WithTimeout(ctx, 2 * time.Second)
			if err := repo.CreateOrder(child, data); err != nil {
				log.Printf("err: %s", err)
			}
			cancel()
		}
	}
	
}
