package models

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

// Db access the underlying db
type Db struct {
	client   *mongo.Client
	database *mongo.Database
}

// Open a new connection to the db and sets the the client
func (d *Db) Open(uri string) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	d.client = client
	d.database = client.Database("kaboo")
	log.Printf("Connected to MongoDB (%v)\n", uri)
}
