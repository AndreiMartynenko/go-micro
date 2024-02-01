package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer used for receiving events from the queue
type Consumer struct {
	conn *amqp.Connection
	//what queue we are going to be dealing with
	queueName string
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}
	//before I return anything I have to set this up
	err := consumer.setup()
	if err != nil {
		//if it's an err I just return an empty consumer and err.
		return Consumer{}, err
	}
	// if it worker I return consumer and nil
	return consumer, nil
}

// set up this consumer. We're going to open up a channel and declare an exchange specific to AMQP protocol

func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return nil
	}
	//We want to return the result of declaring and echange because that's what we need to do here
	// We have to get a channel and exchange
	return declareExchange(channel)
}

//Pushing events to RabbitMQ

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

//func that Listens to the queue and listens for specific topics

func (consumer *Consumer) Listen(topics []string) error {
	//
	ch, err := consumer.conn.Channel()
	// if we can't get the channel
	if err != nil {
		return err
	}
	// we have a channel and we want to close it when we done with it
	//otherwise resource leak
	defer ch.Close()

	//Now we need to get a random queue
	//Common way of working with RabbitMQ

	q, err := declareRandomQueue(ch)
	if err != nil {
		return err
	}

	for _, s := range topics {
		//bind our channel to each of these topics
		ch.QueueBind(
			q.Name,
			s,
			"logs_topic",
			false,
			nil,
		)
		// inside this loop check the error
		if err != nil {
			return err
		}
	}
	//look for messages
	//1. q.Name - what to consume
	//2. "" - consumer
	//3. autoacknowledge - true
	//4. is it exclusive - false
	//5. is it no local(internal) - false
	//6. no wait - false
	//7.
	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	// I want consume all of the things that come from RabbitMQ until I exit
	// declare a channel
	forever := make(chan bool)
	//keeps running on own goroutine
	go func() {
		for d := range messages {
			var payload Payload
			//ignore the error
			//I want to read my JSON in that variable
			// current iteration of our messages the Body Unmarshalled into payload
			_ = json.Unmarshal(d.Body, &payload)

			go handlePayload(payload)
		}
	}()

	fmt.Printf("Waiting for message on [Exchange, Queue] [logs_topic, %s]\n", q.Name)
	<-forever

	return nil
}

func handlePayload(payload Payload) {
	switch payload.Name {
	case "log", "event":
		// log whatever we get
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	case "auth":
		//authenticate

		// you can have as many cases as you want, as long as you write the logic
	default:
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}

	}
}

func logEvent(entry Payload) error {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		return err
	}
	return nil

	// app.writeJSON(w, http.StatusAccepted, payload)

}
