package consumer

import (
	"encoding/json"
	"fmt"
	"log"
	"notification/socket/ws"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

func Worker(hub *ws.Hub, conn *amqp.Connection) {

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"websocket",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
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

	err = ch.QueueBind(
		q.Name,
		"ws",
		"notifications",
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
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	go hub.Run()

	for msg := range msgs {

			var queueMessage map[string]interface{}
			err := json.Unmarshal(msg.Body, &queueMessage)
			if err != nil {
				log.Printf("Failed to unmarshal JSON: %v", err)
				continue
			}
			
			message := new(ws.Message)
			
			switch queueMessage["customer_id"].(string) {
			case "*":
				message.CustomerID = nil 
				message.Content = queueMessage["content"].(string)
			default:
				ID ,err := uuid.Parse(queueMessage["customer_id"].(string))
				
				if err != nil {
					log.Printf("uuid: %v", err)
					continue
				}
				message.CustomerID = &ID 
				message.Content = queueMessage["content"].(string)
			}

			hub.Emitter <- message
			fmt.Printf("Received a message: %+v\n", queueMessage)
	}
	
}