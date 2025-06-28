package worker

import (
	"context"
	"image-processing-system/internal/domain"
	"image-processing-system/internal/metadata"
	"image-processing-system/internal/processor"
	"image-processing-system/internal/storage"
	"image-processing-system/pkg/message"
	"log"
	"sync"

	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func ConsumeQueue(ch *amqp.Channel) {
	msgs, _ := ch.Consume("image.urls", "", true, false, false, false, nil)
	sem := make(chan struct{}, 5) // concurrency control
	var wg sync.WaitGroup

	for msg := range msgs {
		sem <- struct{}{}
		wg.Add(1)
		go func(m amqp.Delivery) {
			defer wg.Done()

			env, job, err := message.Decode[domain.ImageJob](m.Body)
			if err != nil {
				log.Printf("decoding failed: %v", err)
				return
			}

			ctx, span := otel.Tracer("worker").Start(context.Background(), "handleJob")
			span.SetAttributes(attribute.String("trace_id", env.TraceID))
			defer span.End()

			for _, url := range job.URLs {
				img, filename, err := processor.DownloadImage(ctx, url)
				if err != nil {
					log.Printf("download error: %v", err)
					continue
				}
				procImg := processor.Grayscale(img)
				_ = storage.UploadToMinio(ctx, procImg)

				result := metadata.ImageProcessedPayload{
					SourceURL: url,
					S3Path:    "s3://images/" + filename,
					Status:    "success",
					TraceID:   env.TraceID,
				}

				msg, _ := message.Encode(env.TraceID, "image-fetcher", result)
				ch.Publish("", "image.processed", false, false, amqp.Publishing{
					ContentType: "application/json",
					Body:        msg,
				})

			}
			<-sem
		}(msg)
	}
	wg.Wait()
}
