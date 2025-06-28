package rabbitmq

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Connect() (*amqp.Connection, *amqp.Channel) {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@rabbitmq:5672/"
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("RabbitMQ connect fail: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel fail: %v", err)
	}

	// Declare queues
	ch.QueueDeclare("image.urls", false, false, false, false, nil)
	ch.QueueDeclare("image.processed", false, false, false, false, nil)

	return conn, ch
}
