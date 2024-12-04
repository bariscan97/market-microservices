package main

import (
	"auth-service/controller"
	"auth-service/database"
	"auth-service/service"
	"auth-service/events/consumer"
	"auth-service/events/publisher"
	"log"
	"os"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
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
		os.Getenv("RABBITMQ_PORT"))

	conn, err := amqp.Dial(connString)

	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}()

	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	
	publish := &publisher.Producer{
		Conn : conn,
	}

	db := database.GetDB()

	authService := service.NewAuthService(db, publish)

	authController := controller.NewAuthController(authService)

	app := fiber.New()

	auth := app.Group("/") 
	{
		auth.Post("/register", authController.Register)
		auth.Post("/login", authController.Login)
		auth.Get("/verify", authController.RegisterVerify)
		auth.Post("/forgotpassword", authController.ForgotPassword)
		auth.All("/resetpassword", authController.ResetPassword)
	}

	go consumer.Consume(authService, conn)

	app.Listen(":" + os.Getenv("PORT"))
}