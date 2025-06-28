package main

import (
	"context"
	"image-processing-system/internal/config"
	"image-processing-system/internal/worker"
	"image-processing-system/pkg/rabbitmq"
	"image-processing-system/pkg/tracing"
	"log"
)

func main() {
	// Load configuration
	cfg := config.LoadImageFetcherConfig()

	// Initialize tracing
	tracer := tracing.Init("image-fetcher")
	defer tracer.Shutdown(context.Background())

	// Connect to RabbitMQ
	conn, ch := rabbitmq.Connect()
	defer conn.Close()
	defer ch.Close()

	// Create and start worker
	imageWorker, err := worker.NewImageWorker(cfg, ch)
	if err != nil {
		log.Fatalf("Failed to create image worker: %v", err)
	}

	log.Println("image-fetcher service starting...")
	imageWorker.Start()
}
