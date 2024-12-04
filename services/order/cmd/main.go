package main

import (
	"order-service/events/consumers"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"github.com/joho/godotenv"
	"order-service/config"
	"order-service/service"
	"os"
	"fmt"
)

func main() {
	err := godotenv.Load("../.env")
	
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	configurationManager  := config.NewConfigurationManager()

	pool := config.GetConnectionPool(configurationManager.PostgreSqlConfig)
	
	orderService := service.NewService(pool)
	
	connString := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASSWORD"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)

	
	RabbitmqConn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	consumers.Consume(RabbitmqConn, orderService)
}