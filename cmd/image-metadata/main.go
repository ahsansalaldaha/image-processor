package main

import (
	"context"
	"image-processing-system/internal/config"
	"image-processing-system/internal/service/metadata"
	"image-processing-system/pkg/rabbitmq"
	"image-processing-system/pkg/tracing"
	"log"
)

func main() {
	// Load configuration
	cfg := config.LoadImageMetadataConfig()

	// Initialize tracing
	tracer := tracing.Init("image-metadata")
	defer tracer.Shutdown(context.Background())

	// Create metadata service
	metadataSvc, err := metadata.NewMetadataService(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to create metadata service: %v", err)
	}

	// Connect to RabbitMQ
	conn, ch := rabbitmq.Connect()
	defer conn.Close()
	defer ch.Close()

	log.Println("image-metadata service consuming processed image queue")
	metadataSvc.ConsumeAndStore(ch)
}
