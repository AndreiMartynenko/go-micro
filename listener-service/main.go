package main

func main() {

	//try to connect to rabbitmq

	//start listening for maessages
	//What happens. This appli. is not going to periodically connect to the queue and listen
	//for things that way
	//Instead the queue will push it right to us
	//So we'll listen to certain queues
	//And any time there's a event there, we actually get it directly from the queue

	// create consumer
	// consumer consumes messages from the queue

	// watch the queue and consume events

}
