package service

import (
	"context"
	"customer-service/model"
	"encoding/json"
	"log"
	"time"
	"fmt"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type CustomerService struct {
	collection *mongo.Collection
	rabbitConn *amqp.Connection
}

type ICustomerService interface {
	GetCustomerById(id uuid.UUID) (*model.CustomerModel, error)
	CreateCustomer(data *model.CustomerModel) error
	DeleteCustomerById(id uuid.UUID) error
	AddCustomerFieldById(id uuid.UUID, data *model.CustomerModel) error
}

func NewCustomerService(collection *mongo.Collection, rabbitConn *amqp.Connection) ICustomerService {
	return &CustomerService{
		collection: collection,
		rabbitConn:rabbitConn,
	}
}

func (customerService *CustomerService) CreateCustomer(data *model.CustomerModel) error {
	_, err := customerService.collection.InsertOne(context.TODO(), data)
	if err != nil {
		return err
	}	
	return nil
}

func (customerService *CustomerService) GetCustomerById(id uuid.UUID) (*model.CustomerModel, error) {
	var user model.CustomerModel
	if err := customerService.collection.FindOne(context.TODO(), bson.M{"customer_id": id}).Decode(&user); err != nil {
		return &model.CustomerModel{}, err
	}
	return &user, nil
}

func (customerService *CustomerService) DeleteCustomerById(id uuid.UUID) error {	
	result, err := customerService.collection.DeleteOne(context.TODO(), bson.M{"customer_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return mongo.ErrNilValue
	}
	customerService.PublishUserDeletedEvent(id)
	time.Sleep(time.Second * 5)
	return nil
}

func (customerService *CustomerService) AddCustomerFieldById(id uuid.UUID, data *model.CustomerModel) error {
	update := bson.M{"$set": data}
	result, err := customerService.collection.UpdateOne(context.TODO(), bson.M{"customer_id": id}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		data.CustomerId = id 
		err := customerService.CreateCustomer(data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (customerService *CustomerService) PublishUserDeletedEvent(id uuid.UUID) error {
	
	channel, err := customerService.rabbitConn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	body, err := json.Marshal(map[string]interface{}{"id":id})
	if err != nil {
		return err
	}

	if err := channel.ExchangeDeclare(
		"user-delete-event", 
		"direct",       
		true,          
		false,         
		false,         
		false,         
		nil,           
	); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	if err := channel.PublishWithContext(context.Background(),
		"user-delete-event", 
		"", 
		false,          
		false,          
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	); err != nil {
		return err
	}

	log.Printf("UserDeleted event published: %s", string(body))
	return nil
}