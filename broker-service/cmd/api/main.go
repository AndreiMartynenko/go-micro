package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "80"

type Config struct {
	Rabbit *amqp.Connection
}

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

	app := Config{
		Rabbit: rabbitConn,
	}

	log.Printf("Starting broker service on port %s\n", webPort)

	// define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start server

	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
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
