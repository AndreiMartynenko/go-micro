package main

import (
	"context"
	"log-service/data"
	"log-service/logs"
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
}
