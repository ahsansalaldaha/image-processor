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

	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
)

// AMQPChannelAdapter adapts amqp.Channel to implement ChannelInterface
type AMQPChannelAdapter struct {
	*amqp.Channel
}

func (a *AMQPChannelAdapter) IsClosed() bool {
	return a.Channel.IsClosed()
}

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

	// Create adapter for the channel
	channelAdapter := &AMQPChannelAdapter{Channel: ch}

	// Start metrics server if enabled
	if cfg.Metrics.Enabled {
		go func() {
			mux := http.NewServeMux()
			mux.Handle(cfg.Metrics.Path, promhttp.Handler())
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"status":"healthy","service":"url-ingestor"}`))
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

	// Create router with middleware
	router := handler.NewRouter(channelAdapter)

	// Add middleware - ensure metrics endpoint is accessible
	handler := middleware.LoggingMiddleware(router)
	handler = middleware.CORSMiddleware(handler)

	// Create server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: handler,
	}

	log.Printf("url-ingestor listening on :%s", cfg.Server.Port)
	log.Printf("Available endpoints:")
	log.Printf("  - POST /submit (submit images)")
	log.Printf("  - GET /health (health check)")
	log.Printf("  - GET /status (service status)")
	log.Printf("  - GET /queue/status (queue status)")
	log.Printf("  - GET /stats (system stats)")
	log.Printf("  - GET /metrics (Prometheus metrics)")

	if cfg.Metrics.Enabled {
		log.Printf("Metrics server available on :%s%s", cfg.Metrics.Port, cfg.Metrics.Path)
	}

	log.Fatal(srv.ListenAndServe())
}
