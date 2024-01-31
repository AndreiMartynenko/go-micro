package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	//amqp - Advanced Messaging Queue Protocol
	//This package amqp091-go replaces a community developed pack.
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	//try to connect to rabbitmq
	rabbitConn, err := connect()

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitConn.Close()
	fmt.Println("Connected to RabbitMQ")

	//start listening for messages
	//What happens. This appli. is not going to periodically connect to the queue and listen
	//for things that way
	//Instead the queue will push it right to us
	//So we'll listen to certain queues
	//And any time there's a event there, we actually get it directly from the queue

	// create consumer
	// consumer consumes messages from the queue

	// watch the queue and consume events

}

func connect() (*amqp.Connection, error) {
	//this we're going to return once we successfully connect
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	//don't continue until rabbit is ready
	for {
		c, err := amqp.Dial("amqp://guest:geust@localhost") //login and password
		if err != nil {
			fmt.Println("RabbitMQ not yet ready...")
			counts++
		} else {
			// if i connect break the loop
			connection = c
			break
		}
		// if we can't connect after five tries, something is wrong
		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off...")
		time.Sleep(backOff)
		continue

	}

	return connection, nil
}
