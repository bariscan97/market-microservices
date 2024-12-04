package main

import (
	"context"
	"customer-service/config"
	controller "customer-service/http-handler"
	"customer-service/service"
	"customer-service/grpc/handler"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017") )
	
	defer client.Disconnect(context.Background())
	
	if err != nil {
		log.Fatal(err)
	}
	
	RMQconn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	
	

	db := config.GetMongoCollection(client)
	customerService := service.NewCustomerService(db, RMQconn)
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
