package database

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	once        sync.Once
)

func MongoInit() {
	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		clientOpts := options.Client().ApplyURI(os.Getenv("URI"))
		client, err := mongo.Connect(ctx, clientOpts)
		if err != nil {
			log.Fatal("Error connecting to mongo database: ", err)
		}

		if err := client.Ping(ctx, nil); err != nil {
			log.Fatal("Error pinging mongo database: ", err)
		}

		mongoClient = client
		log.Println("Connected to mongo database")
	})
}

func MongoGetClient() *mongo.Client {
	if mongoClient == nil {
		log.Fatal("MongoDB is not initialized. Call MongoInit() first.")
	}
	return mongoClient
}

func MongoGetCollection(databaseName, collectionName string) *mongo.Collection {
	return MongoGetClient().Database(databaseName).Collection(collectionName)
}
