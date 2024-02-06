package main

import (
	"context"
	"fmt"
	"log"
	"log-service/data"
	"log-service/logs"
	"net"

	"google.golang.org/grpc"
)

type LogServer struct {
	// this type of a struct is going to be required for every service ever write over gRPC
	//it's ensure backwards compatibility
	logs.UnimplementedLogServiceServer
	//to have an access to the necessary methods to write to Mongo
	Models data.Models
}

func (l *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	// it gets an Input
	input := req.GetLogEntry()

	//write the log
	logEntry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}

	err := l.Models.LogEntry.Insert(logEntry)
	if err != nil {
		res := &logs.LogResponse{Result: "failed"}
		return res, err
	}

	// return response
	res := &logs.LogResponse{Result: "logged!"}
	return res, nil
}

// create listener
func (app *Config) gRPCListen() {
	lis, err := net.Listen("tpc", fmt.Sprintf(":%s", gRpcPort))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)

		s := grpc.NewServer()

		logs.RegisterLogServiceServer(s, &LogServer{Models: app.Models})
		log.Printf("gRPC Server started on port %s", gRpcPort)

		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to listen for gRPC: %v", err)

		}
	}
}
