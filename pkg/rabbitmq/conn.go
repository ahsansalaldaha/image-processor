package rabbitmq

import (
	"log"

	"github.com/streadway/amqp"
)

func Connect() (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatalf("RabbitMQ connect fail: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("channel fail: %v", err)
	}
	ch.QueueDeclare("image.urls", false, false, false, false, nil)
	return conn, ch
}
