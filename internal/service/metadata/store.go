package metadata

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"image-processing-system/internal/config"
	"image-processing-system/pkg/message"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	recordsStored = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "records_stored_total",
			Help: "Total number of records stored in database",
		},
		[]string{"status"},
	)

	storageDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "storage_duration_seconds",
			Help:    "Database storage operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	dbConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)
)

func init() {
	prometheus.MustRegister(recordsStored)
	prometheus.MustRegister(storageDuration)
	prometheus.MustRegister(dbConnections)
}

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
	db            *gorm.DB
	metricsServer *http.Server
}

// NewMetadataService creates a new metadata service instance
func NewMetadataService(cfg config.DatabaseConfig) (*MetadataService, error) {
	// Use a more compatible connection string format for PostgreSQL 17
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=10",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Add connection pool settings for better stability
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get the underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto migrate the schema
	if err := db.AutoMigrate(&ImageRecord{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Start metrics server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy","service":"image-metadata"}`))
	})

	metricsServer := &http.Server{
		Addr:    ":8083",
		Handler: mux,
	}

	go func() {
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	return &MetadataService{db: db, metricsServer: metricsServer}, nil
}

// ConsumeAndStore processes messages and stores metadata
func (m *MetadataService) ConsumeAndStore(ch *amqp.Channel) {
	msgs, err := ch.Consume("image.processed", "", true, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to consume messages: %v", err)
		return
	}

	for msg := range msgs {
		start := time.Now()

		env, payload, err := message.Decode[ImageProcessedPayload](msg.Body)
		if err != nil {
			log.Printf("Failed to decode message: %v", err)
			recordsStored.WithLabelValues("decode_error").Inc()
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
			recordsStored.WithLabelValues("error").Inc()
		} else {
			log.Printf("Saved image record: %s -> %s", payload.SourceURL, payload.S3Path)
			recordsStored.WithLabelValues("success").Inc()
		}

		storageDuration.Observe(time.Since(start).Seconds())
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
