package main

import (
	"context"
	"image-processing-system/internal/config"
	"image-processing-system/internal/handler"
	"image-processing-system/internal/middleware"
	"image-processing-system/pkg/rabbitmq"
	"image-processing-system/pkg/tracing"
	"log"
	"net/http"
)

func main() {
	// Load configuration
	cfg := config.LoadURLIngestorConfig()

	// Initialize tracing
	tracer := tracing.Init("url-ingestor")
	defer tracer.Shutdown(context.Background())

	// Connect to RabbitMQ
	conn, ch := rabbitmq.Connect()
	defer conn.Close()
	defer ch.Close()

	// Create router with middleware
	router := handler.NewRouter(ch)

	// Add middleware
	handler := middleware.LoggingMiddleware(router)
	handler = middleware.CORSMiddleware(handler)

	// Create server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: handler,
	}

	log.Printf("url-ingestor listening on :%s", cfg.Server.Port)
	log.Fatal(srv.ListenAndServe())
}
