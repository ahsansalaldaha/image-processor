package metadata

import (
	"log"
	"time"

	"image-processing-system/pkg/message"

	"github.com/streadway/amqp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ImageRecord struct {
	ID          uint `gorm:"primaryKey"`
	SourceURL   string
	S3Path      string
	ProcessedAt time.Time
	Status      string
	ErrorMsg    string
	TraceID     string
}

type ImageProcessedPayload struct {
	SourceURL string `json:"source_url"`
	S3Path    string `json:"s3_path"`
	Status    string `json:"status"` // success/error
	ErrorMsg  string `json:"error_msg,omitempty"`
	TraceID   string `json:"trace_id"`
}

func InitDB() *gorm.DB {
	dsn := "host=postgres user=postgres password=postgres dbname=images port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB init error: %v", err)
	}
	db.AutoMigrate(&ImageRecord{})
	return db
}

func ConsumeAndStore(ch *amqp.Channel, db *gorm.DB) {
	msgs, _ := ch.Consume("image.processed", "", true, false, false, false, nil)

	for msg := range msgs {
		env, payload, err := message.Decode[ImageProcessedPayload](msg.Body)
		if err != nil {
			log.Printf("decode failed: %v", err)
			continue
		}

		rec := ImageRecord{
			SourceURL:   payload.SourceURL,
			S3Path:      payload.S3Path,
			ProcessedAt: env.Timestamp,
			Status:      payload.Status,
			ErrorMsg:    payload.ErrorMsg,
			TraceID:     payload.TraceID,
		}
		if err := db.Create(&rec).Error; err != nil {
			log.Printf("DB save failed: %v", err)
		}
	}
}
