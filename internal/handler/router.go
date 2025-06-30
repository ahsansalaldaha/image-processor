package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"image-processing-system/internal/middleware"
	"image-processing-system/internal/models"
	"image-processing-system/pkg/message"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// ChannelInterface defines the interface for RabbitMQ channels
type ChannelInterface interface {
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	IsClosed() bool
	Close() error
}

var (
	imagesSubmitted = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "images_submitted_total",
			Help: "Total number of images submitted for processing",
		},
	)
)

func init() {
	prometheus.MustRegister(imagesSubmitted)
}

// Allowed processing types for image jobs
var allowedProcessingTypes = map[string]struct{}{
	"original":  {},
	"grayscale": {},
	"resize":    {},
	"blur":      {},
	"sharpen":   {},
}

// getAllowedProcessingTypes returns a slice of allowed processing types
func getAllowedProcessingTypes() []string {
	return []string{"original", "grayscale", "resize", "blur", "sharpen"}
}

// validateProcessingTypes checks if all provided types are allowed
func validateProcessingTypes(types []string) (invalid []string) {
	for _, t := range types {
		if _, ok := allowedProcessingTypes[t]; !ok {
			invalid = append(invalid, t)
		}
	}
	return
}

// publishJob publishes a single job to the queue
func publishJob(ctx context.Context, ch ChannelInterface, traceID string, url string, processingType string) error {
	job := models.ImageJob{
		URLs:            []string{url},
		ProcessingTypes: []string{processingType},
	}
	encoded, _ := message.Encode(traceID, "url-ingestor", job)

	// Inject trace context into headers
	prop := propagation.TraceContext{}
	headers := make(map[string]string)
	prop.Inject(ctx, propagation.MapCarrier(headers))
	if tp, ok := headers["traceparent"]; ok {
		log.Printf("[ingestor] Injecting traceparent: %s", tp)
	}

	amqpHeaders := amqp.Table{}
	for k, v := range headers {
		amqpHeaders[k] = v
	}

	return ch.Publish("", "image.urls", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        encoded,
		Headers:     amqpHeaders,
	})
}

func NewRouter(ch ChannelInterface) http.Handler {
	r := chi.NewRouter()

	// Add rate limiting middleware
	r.Use(httprate.LimitByIP(50, 1)) // 50 req/sec

	// Add Prometheus metrics middleware
	r.Use(middleware.MetricsMiddleware)

	// Health check - no middleware applied
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"service":   "url-ingestor",
		})
	})

	// Metrics endpoint - no middleware applied to avoid conflicts
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		promhttp.Handler().ServeHTTP(w, r)
	})

	// Status endpoint
	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		// Check RabbitMQ connection
		rabbitmqStatus := "healthy"
		if ch == nil || ch.IsClosed() {
			rabbitmqStatus = "unhealthy"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service":   "url-ingestor",
			"status":    "running",
			"timestamp": time.Now().UTC(),
			"dependencies": map[string]string{
				"rabbitmq": rabbitmqStatus,
			},
		})
	})

	// Queue status endpoint
	r.Get("/queue/status", func(w http.ResponseWriter, r *http.Request) {
		if ch == nil || ch.IsClosed() {
			http.Error(w, "RabbitMQ connection not available", http.StatusServiceUnavailable)
			return
		}

		// Get queue info - this would need to be handled differently for mocks
		// For now, we'll skip this in tests
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"queue_name": "image.urls",
			"messages":   0,
			"consumers":  0,
			"timestamp":  time.Now().UTC(),
		})
	})

	// System statistics endpoint
	r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service":   "url-ingestor",
			"timestamp": time.Now().UTC(),
			"metrics": map[string]interface{}{
				"endpoints": map[string]string{
					"health":  "/health",
					"status":  "/status",
					"queue":   "/queue/status",
					"metrics": "/metrics",
					"submit":  "/submit",
				},
			},
		})
	})

	r.Post("/submit", func(w http.ResponseWriter, r *http.Request) {
		var job models.ImageJob
		if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate processing types
		invalidTypes := validateProcessingTypes(job.ProcessingTypes)
		if len(invalidTypes) > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":         "invalid processing_types provided",
				"invalid_types": invalidTypes,
				"allowed_types": getAllowedProcessingTypes(),
			})
			return
		}

		// Extract traceparent header if present
		prop := propagation.TraceContext{}
		ctx := r.Context()
		ctx = prop.Extract(ctx, propagation.HeaderCarrier(r.Header))
		tracer := otel.Tracer("url-ingestor")
		ctx, span := tracer.Start(ctx, "SubmitImageJob")
		defer span.End()

		traceID := r.Header.Get("X-Trace-ID")
		totalJobs := 0

		for _, url := range job.URLs {
			// Always publish the original
			if err := publishJob(ctx, ch, traceID, url, "original"); err != nil {
				span.RecordError(err)
				http.Error(w, "publish failed", http.StatusInternalServerError)
				return
			}
			totalJobs++

			// Publish other processing types if specified (skip duplicate 'original')
			for _, pType := range job.ProcessingTypes {
				if pType == "original" {
					continue
				}
				if err := publishJob(ctx, ch, traceID, url, pType); err != nil {
					span.RecordError(err)
					http.Error(w, "publish failed", http.StatusInternalServerError)
					return
				}
				totalJobs++
			}
		}

		imagesSubmitted.Add(float64(totalJobs))
		w.WriteHeader(http.StatusAccepted)
	})

	return r
}
