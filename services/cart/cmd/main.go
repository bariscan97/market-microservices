package main

import (
	"cart-service/events/consumers/broker"
	customer_pb "cart-service/grpc/pb/customer-rpc"
	product_pb "cart-service/grpc/pb/product-rpc"
	"cart-service/handler"
	"cart-service/events/publishers"
	"cart-service/service"
	"context"
	"log"
	"net/http"
	"os"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	
	err := godotenv.Load()
	
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	addr := fmt.Sprintf(
		"%s:%s",
		os.Getenv("REDIS_HOST"),
		os.Getenv("REDIS_PORT"),
	)

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

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
	
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	prodcutDial, err :=  grpc.NewClient(":55002", opts...)

	if err != nil {
		log.Fatalf("failed to connect to server: %v", err)
	}

	customerDial, err := grpc.NewClient(":55001", opts...)
	
	if err != nil {
		log.Fatalf("failed to connect to server: %v", err)
	}
	
	defer func ()  {
		RabbitmqConn.Close()
		client.Close()
		customerDial.Close()
	}()	
		

	gprcProductClient := product_pb.NewProductServiceClient(prodcutDial)
	grpcCustomerClient := customer_pb.NewCustomerServiceClient(customerDial)
	
	producer := publishers.NewProducer(RabbitmqConn, context.Background())
	
	cartService := service.NewCacheClient(
		client, 
		producer.Notify, 
		producer.PushOrder, 
		grpcCustomerClient,
		gprcProductClient,
	)
	cartHandler := handler.NewCustomerController(cartService)
	consumers := broker.NewBroker(cartService, RabbitmqConn)
	
	r := mux.NewRouter()

	r.HandleFunc("/",cartHandler.AddItemToCart).Methods("POST")
	r.HandleFunc("/",cartHandler.GetCartByCustomerId).Methods("GET")
	r.HandleFunc("/buy",cartHandler.Buy).Methods("GET")
	
	go producer.Worker()
	go consumers.Run()
	
	http.Handle("/", r)
	if err := http.ListenAndServe(":8083", nil); err != nil {
		panic(err)
	}
}
