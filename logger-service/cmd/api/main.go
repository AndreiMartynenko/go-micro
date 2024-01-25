package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	//because we're running a Docker the port would be 80
	webPort  = "80"
	rpcPort  = "5001"
	mongoURL = "mongodb://mongo:27017"
	gRpcPort = "50001"
)

// mongo client as a package level variable
var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	//connect to mongo
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}
	client = mongoClient

	// create a context in order to disconnect Mongo

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	//close connection
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}

}

func (app *Config) serve() {
	//Starts a web server
	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", webPort),
	}
}

// func to connect the mongo
func connectToMongo() (*mongo.Client, error) {
	/*
		Create Connection Options:
		Utilizes the options package to create connection options.
		Applies the MongoDB connection URI (mongoURL variable) to the options.
	*/
	clientOptions := options.Client().ApplyURI(mongoURL)

	/*
		Sets the authentication credentials (username and password) for MongoDB.
	*/
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	/*
		Establish Connection:
		Uses the mongo.Connect function to establish a connection to MongoDB.
		Uses the provided context (context.TODO()).
	*/
	c, err := mongo.Connect(context.TODO(), clientOptions)

	/*
		Handle Connection Errors:
		If an error occurs during the connection process,
		logs the error and returns nil client and the encountered error.
	*/
	if err != nil {
		log.Println("Error connecting:", err)
		return nil, err
	}
	/*
			Return MongoDB Client:

		If the connection is successful, returns the MongoDB client (c) and nil error.

	*/

	return c, nil

}
