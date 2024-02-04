package main

import (
	"context"
	"log"
	"log-service/data"
	"time"
)

// first thing we do is we declare a type that's going to be the RPCServer
type RPCServer struct {
}

// then we declare a type that's the kind of data we're going to receive
// for any methods that are tried to our RPCServer.
type RPCPayload struct {
	Name string
	Data string
}

func (r *RPCServer) LogInfo(payload RPCPayload, resp *string) error {
	collection := client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), data.LogEntry{
		Name:      payload.Name,
		Data:      payload.Data,
		CreatedAt: time.Now(),
	})
	if err != nil {
		log.Println("error writing to mongo", err)
		return err
	}

	*resp = "Processed payload via RPC:" + payload.Name
	return nil
}
