package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"image-processing-system/internal/config"
	"image-processing-system/internal/middleware"
	"image-processing-system/internal/models"
	"image-processing-system/internal/service/metadata"
	"image-processing-system/internal/service/processor"
	"image-processing-system/internal/service/storage"
	"image-processing-system/pkg/message"

	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// ImageWorker handles image processing jobs
type ImageWorker struct {
	config           *config.ImageFetcherConfig
	processor        *processor.ImageProcessor
	storage          *storage.MinioService
	metadata         *metadata.MetadataService
	channel          *amqp.Channel
	concurrencyLimit int
	metricsServer    *http.Server
}

// NewImageWorker creates a new image worker instance
func NewImageWorker(cfg *config.ImageFetcherConfig, ch *amqp.Channel) (*ImageWorker, error) {
	proc := processor.NewImageProcessor()

	storageSvc, err := storage.NewMinioService(cfg.Minio)
	if err != nil {
		return nil, err
	}

	metadataSvc, err := metadata.NewMetadataService(cfg.Database)
	if err != nil {
		return nil, err
	}

	// Start metrics server if enabled
	var metricsServer *http.Server
	if cfg.Metrics.Enabled {
		mux := http.NewServeMux()
		mux.Handle(cfg.Metrics.Path, promhttp.Handler())
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"healthy","service":"image-fetcher"}`))
		})

		metricsServer = &http.Server{
			Addr:    ":" + cfg.Metrics.Port,
			Handler: mux,
		}

		go func() {
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("Metrics server error: %v", err)
			}
		}()
	}

	return &ImageWorker{
		config:           cfg,
		processor:        proc,
		storage:          storageSvc,
		metadata:         metadataSvc,
		channel:          ch,
		concurrencyLimit: 5, // Can be made configurable
		metricsServer:    metricsServer,
	}, nil
}

// Start begins consuming and processing image jobs
func (w *ImageWorker) Start() {
	msgs, err := w.channel.Consume("image.urls", "", true, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to consume messages: %v", err)
		return
	}

	sem := make(chan struct{}, w.concurrencyLimit)
	var wg sync.WaitGroup

	for msg := range msgs {
		sem <- struct{}{}
		wg.Add(1)
		middleware.ActiveWorkers.WithLabelValues("image-fetcher").Inc()

		go func(m amqp.Delivery) {
			defer wg.Done()
			defer func() {
				<-sem
				middleware.ActiveWorkers.WithLabelValues("image-fetcher").Dec()
			}()

			w.processJob(m)
		}(msg)
	}
	wg.Wait()
}

// processJob processes a single image job
func (w *ImageWorker) processJob(msg amqp.Delivery) {
	start := time.Now()

	env, job, err := message.Decode[models.ImageJob](msg.Body)
	if err != nil {
		log.Printf("Failed to decode job: %v", err)
		middleware.JobsProcessed.WithLabelValues("decode_error", "image-fetcher").Inc()
		return
	}

	ctx, span := otel.Tracer("worker").Start(context.Background(), "processJob")
	span.SetAttributes(attribute.String("trace_id", env.TraceID))
	defer span.End()

	successCount := 0
	errorCount := 0

	for _, url := range job.URLs {
		if err := w.processImage(ctx, url, env.TraceID); err != nil {
			log.Printf("Failed to process image %s: %v", url, err)
			errorCount++
			// Continue processing other images in the job
		} else {
			successCount++
		}
	}

	// Record metrics
	middleware.ImagesProcessed.WithLabelValues("success", "image-fetcher").Add(float64(successCount))
	middleware.ImagesProcessed.WithLabelValues("error", "image-fetcher").Add(float64(errorCount))
	middleware.JobProcessingDuration.WithLabelValues("image-fetcher").Observe(time.Since(start).Seconds())
}

// processImage processes a single image
func (w *ImageWorker) processImage(ctx context.Context, url, traceID string) error {
	// Download image
	downloadStart := time.Now()
	img, _, err := w.processor.DownloadImage(ctx, url)
	if err != nil {
		middleware.ProcessingDuration.WithLabelValues("download", "image-fetcher").Observe(time.Since(downloadStart).Seconds())
		return err
	}
	middleware.ProcessingDuration.WithLabelValues("download", "image-fetcher").Observe(time.Since(downloadStart).Seconds())

	// Process image (convert to grayscale)
	processStart := time.Now()
	processedImg := w.processor.Grayscale(img)
	middleware.ProcessingDuration.WithLabelValues("grayscale", "image-fetcher").Observe(time.Since(processStart).Seconds())

	// Upload to storage
	uploadStart := time.Now()
	filename, err := w.storage.UploadImage(ctx, processedImg)
	if err != nil {
		middleware.ProcessingDuration.WithLabelValues("upload", "image-fetcher").Observe(time.Since(uploadStart).Seconds())
		return err
	}
	middleware.ProcessingDuration.WithLabelValues("upload", "image-fetcher").Observe(time.Since(uploadStart).Seconds())

	// Create result payload
	result := metadata.ImageProcessedPayload{
		SourceURL: url,
		S3Path:    w.storage.GetImageURL(filename),
		Status:    "success",
		TraceID:   traceID,
	}

	// Publish result
	encoded, err := message.Encode(traceID, "image-fetcher", result)
	if err != nil {
		return err
	}

	err = w.channel.Publish("", "image.processed", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        encoded,
	})
	if err != nil {
		return err
	}

	log.Printf("Successfully processed image: %s -> %s", url, result.S3Path)
	return nil
}
