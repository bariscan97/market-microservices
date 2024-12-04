package main

import (
	"catalog-service/cache"
	"catalog-service/consumer"
	"catalog-service/handler"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	err := godotenv.Load("../.env")
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

	conn, err := amqp.Dial(connString)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}

	repo := cache.NewCacheClient(client)

	productHandler := handler.NewCustomerController(repo)

	go consumer.Consumer(repo, conn)

	r := mux.NewRouter() 
	{	r.HandleFunc("/health",func(w http.ResponseWriter, r *http.Request) {
			x := r.Header.Get("nane")
			fmt.Println("hello",x)
			json.NewEncoder(w).Encode("ok")
		})
		r.HandleFunc("/products", productHandler.GetProducts).Methods("GET")
		r.HandleFunc("/products/categories", productHandler.GetGategories).Methods("GET")
		r.HandleFunc("/products/{id}",productHandler.GetProductById).Methods("GET")
		
	}
	
	if err := http.ListenAndServe(":8082", r); err != nil {
		panic(err)
	}
}