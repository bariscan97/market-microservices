package broker

import (
	product_event "cart-service/events/consumers/product-event"
	user_event "cart-service/events/consumers/user-event"
	"cart-service/service"
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Broker struct {
	Conn *amqp.Connection
	Repo *service.RedisClient
	UserEvent func(repo *service.RedisClient, conn *amqp.Connection) error
	ProductEvent func(repo *service.RedisClient, conn *amqp.Connection) error
}

func NewBroker(repo *service.RedisClient, conn *amqp.Connection) *Broker {
	return &Broker{
		Conn: conn,
		Repo: repo,
		UserEvent: user_event.Consume,
		ProductEvent: product_event.Consume,
	}
}

func (broker *Broker) Run() {
	go broker.Repo.EventListener(context.Background())
	go broker.UserEvent(broker.Repo, broker.Conn)
	go broker.ProductEvent(broker.Repo, broker.Conn)
}

