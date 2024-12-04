package main

import (
	"context"
	"customer-service/config"
	controller "customer-service/http-handler"
	"customer-service/service"
	"customer-service/grpc/handler"
	"log"
	"net/http"
	"os"
	"fmt"
	"github.com/gorilla/mux"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017") )
	
	defer client.Disconnect(context.Background())
	
	if err != nil {
		log.Fatal(err)
	}
	
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
	
	

	db := config.GetMongoCollection(client)
	customerService := service.NewCustomerService(db, RabbitmqConn)
	customerController := controller.NewCustomerController(customerService)

	GrpcServer := handler.NewServer(customerService)

	r := mux.NewRouter()

	r.HandleFunc("/", customerController.DeleteCustomerById).Methods("DELETE")
	r.HandleFunc("/", customerController.GetCustomerById).Methods("GET")
	r.HandleFunc("/", customerController.AddCustomerFieldById).Methods("PUT")

	http.Handle("/", r)
	go http.ListenAndServe(":8086", nil)
	GrpcServer.Run(":55001")
}
