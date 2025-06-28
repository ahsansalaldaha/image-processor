package worker

import (
	"context"
	"log"
	"sync"

	"image-processing-system/internal/config"
	"image-processing-system/internal/models"
	"image-processing-system/internal/service/metadata"
	"image-processing-system/internal/service/processor"
	"image-processing-system/internal/service/storage"
	"image-processing-system/pkg/message"

	"github.com/streadway/amqp"
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

	return &ImageWorker{
		config:           cfg,
		processor:        proc,
		storage:          storageSvc,
		metadata:         metadataSvc,
		channel:          ch,
		concurrencyLimit: 5, // Can be made configurable
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
		go func(m amqp.Delivery) {
			defer wg.Done()
			defer func() { <-sem }()

			w.processJob(m)
		}(msg)
	}
	wg.Wait()
}

// processJob processes a single image job
func (w *ImageWorker) processJob(msg amqp.Delivery) {
	env, job, err := message.Decode[models.ImageJob](msg.Body)
	if err != nil {
		log.Printf("Failed to decode job: %v", err)
		return
	}

	ctx, span := otel.Tracer("worker").Start(context.Background(), "processJob")
	span.SetAttributes(attribute.String("trace_id", env.TraceID))
	defer span.End()

	for _, url := range job.URLs {
		if err := w.processImage(ctx, url, env.TraceID); err != nil {
			log.Printf("Failed to process image %s: %v", url, err)
			// Continue processing other images in the job
		}
	}
}

// processImage processes a single image
func (w *ImageWorker) processImage(ctx context.Context, url, traceID string) error {
	// Download image
	img, _, err := w.processor.DownloadImage(ctx, url)
	if err != nil {
		return err
	}

	// Process image (convert to grayscale)
	processedImg := w.processor.Grayscale(img)

	// Upload to storage
	filename, err := w.storage.UploadImage(ctx, processedImg)
	if err != nil {
		return err
	}

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
