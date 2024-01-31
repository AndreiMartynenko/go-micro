package event

import (
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
	ch, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

}
