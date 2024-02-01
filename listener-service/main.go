package main

import (
	"fmt"
	"listenner/event"
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
		//not executable
		/*
			os.Exit(1): This line immediately terminates the program and exits with a status code of 1.
			In Unix-like operating systems, an exit status of 1 typically indicates that an error
			occurred during program execution. By exiting the program with
			a non-zero status code, it signals to the calling process (e.g., the shell)
			that the program encountered an error.


		*/
		os.Exit(1)
	}
	defer rabbitConn.Close()

	//start listening for messages
	//What happens. This appli. is not going to periodically connect to the queue and listen
	//for things that way
	//Instead the queue will push it right to us
	//So we'll listen to certain queues
	//And any time there's a event there, we actually get it directly from the queue

	log.Println("Listening for and consuming RabbitMQ messages...")

	// create consumer
	// consumer consumes messages from the queue
	consumer, err := event.NewConsumer(rabbitConn)
	if err != nil {
		panic(err)
	}

	// watch the queue and consume events
	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})
	if err != nil {
		log.Println(err)
	}

}

func connect() (*amqp.Connection, error) {
	//this we're going to return once we successfully connect
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	//don't continue until rabbit is ready
	for {
		//c, err := amqp.Dial("amqp://guest:guest@localhost") //login and password
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq") //login and password
		if err != nil {
			fmt.Println("RabbitMQ not yet ready...")
			counts++
		} else {
			// if I connect, break the loop
			fmt.Println("Connected to RabbitMQ")
			connection = c
			break
		}
		// if we can't connect after five tries, something is wrong
		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}
		// if we haven't tried at least five times or at most five times, then I'll back off
		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off...")
		time.Sleep(backOff)
		continue

	}

	return connection, nil
}
