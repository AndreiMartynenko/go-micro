package main

import (
	"context"
	"fmt"
	"log"
	"log-service/data"
	"net"
	"net/http"
	"net/rpc"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	//because we're running a Docker the port would be 80
	webPort  = "80"
	rpcPort  = "5001"
	mongoURL = "mongodb://mongo:27017"
	//mongoURL = "mongodb://localhost:27017"
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
	// Register the RPC Server
	err = rpc.Register(new(RPCServer))
	go app.rpcListen()

	//start web server
	log.Println("Starting service on port", webPort)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Panic()
	}

}

func (app *Config) rpcListen() error {
	log.Println("Starting RPC server on port ", rpcPort)
	//how do we listen on our RPC
	//we declare variable listen
	//we're listening for TCP connections
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", rpcPort))

	if err != nil {
		return err
	}

	defer listen.Close()

	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(rpcConn)
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

	log.Println("Connect to mongo!")
	/*
			Return MongoDB Client:

		If the connection is successful, returns the MongoDB client (c) and nil error.

	*/

	return c, nil

}
