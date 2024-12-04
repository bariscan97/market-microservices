package main

import (
	"log"
	"net/http"
	"notification/socket/consumer"
	"notification/socket/ws"
	"notification/mailer/sender"
	"notification/mailer/worker"
	"github.com/gorilla/mux"
	"os"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
	"fmt"
)

func main() {
	err := godotenv.Load("../.env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	connString := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASSWORD"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)

	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	defer conn.Close()

	r := mux.NewRouter()

	hub := ws.NewHub()
	handler := ws.NewHandler(hub)

	go consumer.Worker(hub, conn)

	mailer := sender.Mailer{
		From :os.Getenv("SMTP_PUBLISHER"),
		Password :os.Getenv("SMTP_PASSWORD"),
		SmtpHost :os.Getenv("SMTP_HOST"),
		SmtpPort :os.Getenv("SMTP_PORT"),
	}

	go worker.MailWorker(&mailer, conn)

	r.HandleFunc("/ws/join", handler.JoinWs)
	
	log.Println("WebSocket server running on :8085")
	if err := http.ListenAndServe(":8085", r); err != nil {
		log.Fatal("HTTP server error:", err)
	}
}
