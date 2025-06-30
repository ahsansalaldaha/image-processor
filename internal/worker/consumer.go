package worker

import (
	"context"
	"fmt"
	"image"
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

	// Each job now contains a single URL and a single processing type
	if len(job.URLs) == 0 || len(job.ProcessingTypes) == 0 {
		log.Printf("Job missing URL or processing type")
		return
	}
	url := job.URLs[0]
	processingType := job.ProcessingTypes[0]

	if err := w.processImage(ctx, url, processingType, env.TraceID); err != nil {
		log.Printf("Failed to process image %s [%s]: %v", url, processingType, err)
		errorCount++
	} else {
		successCount++
	}

	// Record metrics
	middleware.ImagesProcessed.WithLabelValues("success", "image-fetcher").Add(float64(successCount))
	middleware.ImagesProcessed.WithLabelValues("error", "image-fetcher").Add(float64(errorCount))
	middleware.JobProcessingDuration.WithLabelValues("image-fetcher").Observe(time.Since(start).Seconds())
}

// processImage processes a single image with the given processing type
func (w *ImageWorker) processImage(ctx context.Context, url, processingType, traceID string) error {
	// Download image
	downloadStart := time.Now()
	img, format, err := w.processor.DownloadImage(ctx, url)
	if err != nil {
		middleware.ProcessingDuration.WithLabelValues("download", "image-fetcher").Observe(time.Since(downloadStart).Seconds())
		return err
	}
	middleware.ProcessingDuration.WithLabelValues("download", "image-fetcher").Observe(time.Since(downloadStart).Seconds())

	// Extract image dimensions
	width := 0
	height := 0
	if img != nil {
		width = img.Bounds().Dx()
		height = img.Bounds().Dy()
	}

	// Process image according to processingType
	processStart := time.Now()
	var processedImg image.Image
	switch processingType {
	case "original":
		processedImg = img // store as-is
		middleware.ProcessingDuration.WithLabelValues("original", "image-fetcher").Observe(time.Since(processStart).Seconds())
	case "grayscale":
		processedImg = w.processor.Grayscale(img)
		middleware.ProcessingDuration.WithLabelValues("grayscale", "image-fetcher").Observe(time.Since(processStart).Seconds())
	case "resize":
		processedImg = w.processor.Resize(img, 100, 100)
		middleware.ProcessingDuration.WithLabelValues("resize", "image-fetcher").Observe(time.Since(processStart).Seconds())
	case "blur":
		processedImg = w.processor.Blur(img, 2.0)
		middleware.ProcessingDuration.WithLabelValues("blur", "image-fetcher").Observe(time.Since(processStart).Seconds())
	case "sharpen":
		processedImg = w.processor.Sharpen(img, 2.0)
		middleware.ProcessingDuration.WithLabelValues("sharpen", "image-fetcher").Observe(time.Since(processStart).Seconds())
	default:
		return fmt.Errorf("unsupported processing type: %s", processingType)
	}

	// Upload to storage (pass processingType for filename)
	uploadStart := time.Now()
	filename, err := w.storage.UploadImageWithType(ctx, processedImg, processingType)
	if err != nil {
		middleware.ProcessingDuration.WithLabelValues("upload", "image-fetcher").Observe(time.Since(uploadStart).Seconds())
		return err
	}
	middleware.ProcessingDuration.WithLabelValues("upload", "image-fetcher").Observe(time.Since(uploadStart).Seconds())

	// Get file size from MinIO
	fileSize, err := w.storage.GetFileSize(ctx, filename)
	if err != nil {
		log.Printf("Failed to get file size for %s: %v", filename, err)
		fileSize = 0
	}

	// Create result payload
	result := models.ImageProcessedPayload{
		SourceURL:      url,
		S3Path:         w.storage.GetImageURL(filename),
		Status:         "success",
		TraceID:        traceID,
		Width:          width,
		Height:         height,
		Format:         format,
		FileSize:       fileSize,
		ProcessingType: processingType,
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

	log.Printf("Successfully processed image: %s [%s] -> %s", url, processingType, result.S3Path)
	return nil
}
