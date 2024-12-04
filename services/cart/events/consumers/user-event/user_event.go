package user_event

import (
	"cart-service/service"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

func Consume(repo *service.RedisClient, conn *amqp.Connection) error {

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
		return fmt.Errorf("Failed to open a channel: %v", err)
	}

	exchangeName := "user-delete-event" 
	queueName := "cart_user_event"

	defer func() {
		ch.Close()
		conn.Close()
	}()

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

	go repo.EventListener(context.Background())

	for msg := range msgs {
		var data map[string]interface{}
		err := json.Unmarshal(msg.Body, &data)
		if err != nil {
			fmt.Println("error: ", err)
		}

		id, ok := data["id"].(string)
		if ok {
			parsedID, err := uuid.Parse(id)
			if err != nil {
				log.Printf("uuid parse error:%s", err)
			}
			repo.UserEvent <- &parsedID
		}
	}

	select {}
}