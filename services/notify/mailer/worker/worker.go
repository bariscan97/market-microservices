package worker

import (
	"encoding/json"
	"fmt"
	"log"
	mailer "notification/mailer/sender"
	amqp "github.com/rabbitmq/amqp091-go"
)

type EmailMessage struct {
	Template string `json:"template"`
	To       string `json:"to"`
}

func MailWorker(sender *mailer.Mailer, conn *amqp.Connection) error {

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
		return fmt.Errorf("Failed to open a channel: %v", err)
	}
	defer func() {
		ch.Close()
		conn.Close()
	}()

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

	go func() {
		for msg := range msgs {

			var emailMessage EmailMessage

			err := json.Unmarshal(msg.Body, &emailMessage)
			if err != nil {
				log.Printf("Error unmarshalling message: %v", err)
				continue
			}
			sender.SendMail(emailMessage.Template, emailMessage.To)

			log.Printf("Received message: %+v", emailMessage.Template)

		}
	}()

	select {}
}
