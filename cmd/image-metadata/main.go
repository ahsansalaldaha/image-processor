package main

import (
	"context"
	"image-processing-system/internal/config"
	"image-processing-system/internal/service/metadata"
	"image-processing-system/pkg/rabbitmq"
	"image-processing-system/pkg/tracing"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load configuration
	cfg := config.LoadImageMetadataConfig()

	// Initialize tracing
	tracer := tracing.Init("image-metadata")
	defer tracer.Shutdown(context.Background())

	// Start metrics server if enabled
	if cfg.Metrics.Enabled {
		go func() {
			mux := http.NewServeMux()
			mux.Handle(cfg.Metrics.Path, promhttp.Handler())
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"status":"healthy","service":"image-metadata"}`))
			})

			metricsServer := &http.Server{
				Addr:    ":" + cfg.Metrics.Port,
				Handler: mux,
			}

			log.Printf("Metrics server listening on :%s", cfg.Metrics.Port)
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("Metrics server error: %v", err)
			}
		}()
	}

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
	if cfg.Metrics.Enabled {
		log.Printf("Metrics server available on :%s%s", cfg.Metrics.Port, cfg.Metrics.Path)
	}
	metadataSvc.ConsumeAndStore(ch)
}
