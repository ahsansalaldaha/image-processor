package main

import (
	"image-processing-system/internal/worker"
	"image-processing-system/pkg/rabbitmq"
	"image-processing-system/pkg/tracing"
	"log"
)

func main() {
	tracer := tracing.Init("image-fetcher")
	defer tracer.Shutdown()

	conn, ch := rabbitmq.Connect()
	defer conn.Close()
	defer ch.Close()

	log.Println("image-fetcher service consuming queue")
	worker.ConsumeQueue(ch)
}
