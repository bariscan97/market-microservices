package config

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func GetMongoCollection(client *mongo.Client) *mongo.Collection {
	err := client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}

	collection :=  client.Database("mydb").Collection("customers") 

	indexModel := mongo.IndexModel{
		Keys: bson.M{"customer_id": 1}, 
		Options: options.Index().SetUnique(true), 
	}
	
	_, err = collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		log.Fatal(err)
	}
	return collection
}
