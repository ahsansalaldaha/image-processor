package metadata

import (
	"fmt"
	"log"
	"time"

	"image-processing-system/internal/config"
	"image-processing-system/pkg/message"

	"github.com/streadway/amqp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ImageRecord represents an image processing record in the database
type ImageRecord struct {
	ID          uint `gorm:"primaryKey"`
	SourceURL   string
	S3Path      string
	ProcessedAt time.Time
	Status      string
	ErrorMsg    string
	TraceID     string
}

// ImageProcessedPayload represents the payload for processed image messages
type ImageProcessedPayload struct {
	SourceURL string `json:"source_url"`
	S3Path    string `json:"s3_path"`
	Status    string `json:"status"` // success/error
	ErrorMsg  string `json:"error_msg,omitempty"`
	TraceID   string `json:"trace_id"`
}

// MetadataService handles metadata operations
type MetadataService struct {
	db *gorm.DB
}

// NewMetadataService creates a new metadata service instance
func NewMetadataService(cfg config.DatabaseConfig) (*MetadataService, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate the schema
	if err := db.AutoMigrate(&ImageRecord{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &MetadataService{db: db}, nil
}

// ConsumeAndStore processes messages and stores metadata
func (m *MetadataService) ConsumeAndStore(ch *amqp.Channel) {
	msgs, err := ch.Consume("image.processed", "", true, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to consume messages: %v", err)
		return
	}

	for msg := range msgs {
		env, payload, err := message.Decode[ImageProcessedPayload](msg.Body)
		if err != nil {
			log.Printf("Failed to decode message: %v", err)
			continue
		}

		record := ImageRecord{
			SourceURL:   payload.SourceURL,
			S3Path:      payload.S3Path,
			ProcessedAt: env.Timestamp,
			Status:      payload.Status,
			ErrorMsg:    payload.ErrorMsg,
			TraceID:     payload.TraceID,
		}

		if err := m.db.Create(&record).Error; err != nil {
			log.Printf("Failed to save record to database: %v", err)
		} else {
			log.Printf("Saved image record: %s -> %s", payload.SourceURL, payload.S3Path)
		}
	}
}

// GetImageRecords retrieves image records from the database
func (m *MetadataService) GetImageRecords(limit int) ([]ImageRecord, error) {
	var records []ImageRecord
	err := m.db.Order("processed_at DESC").Limit(limit).Find(&records).Error
	return records, err
}

// GetImageRecordByID retrieves a specific image record by ID
func (m *MetadataService) GetImageRecordByID(id uint) (*ImageRecord, error) {
	var record ImageRecord
	err := m.db.First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}
