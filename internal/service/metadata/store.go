package metadata

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"image-processing-system/internal/config"
	"image-processing-system/internal/models"
	"image-processing-system/pkg/message"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
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
	if err := db.AutoMigrate(&models.ImageRecord{}); err != nil {
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

		// Extract trace context from AMQP headers (robust for string and []byte)
		prop := propagation.TraceContext{}
		headers := make(map[string]string)
		for k, v := range msg.Headers {
			switch val := v.(type) {
			case string:
				headers[k] = val
			case []byte:
				headers[k] = string(val)
			}
		}
		if tp, ok := headers["traceparent"]; ok {
			log.Printf("[metadata] Consumed traceparent: %s", tp)
		}
		ctx := context.Background()
		ctx = prop.Extract(ctx, propagation.MapCarrier(headers))

		env, payload, err := message.Decode[models.ImageProcessedPayload](msg.Body)
		if err != nil {
			log.Printf("Failed to decode message: %v", err)
			recordsStored.WithLabelValues("decode_error").Inc()
			continue
		}

		tracer := otel.Tracer("image-metadata")
		spanName := "StoreMetadata/" + payload.ProcessingType
		ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindConsumer))
		span.SetAttributes(
			attribute.String("processing_type", payload.ProcessingType),
			attribute.String("status", payload.Status),
			attribute.String("source_url", payload.SourceURL),
			attribute.String("trace_id", payload.TraceID),
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.destination.name", "image.processed"),
			attribute.String("messaging.operation", "process"),
		)
		defer span.End()

		record := models.ImageRecord{
			SourceURL:      payload.SourceURL,
			S3Path:         payload.S3Path,
			ProcessedAt:    env.Timestamp,
			Status:         payload.Status,
			ErrorMsg:       payload.ErrorMsg,
			TraceID:        payload.TraceID,
			Width:          payload.Width,
			Height:         payload.Height,
			Format:         payload.Format,
			FileSize:       payload.FileSize,
			ProcessingType: payload.ProcessingType,
		}

		// Optional: wrap DB create in a child span
		dbCtx, dbSpan := tracer.Start(ctx, "DBCreate")
		if err := m.db.WithContext(dbCtx).Create(&record).Error; err != nil {
			dbSpan.RecordError(err)
			log.Printf("Failed to save record to database: %v", err)
			recordsStored.WithLabelValues("error").Inc()
		} else {
			log.Printf("Saved image record: %s -> %s", payload.SourceURL, payload.S3Path)
			recordsStored.WithLabelValues("success").Inc()
		}
		dbSpan.End()

		storageDuration.Observe(time.Since(start).Seconds())
	}
}

// GetImageRecords retrieves image records from the database
func (m *MetadataService) GetImageRecords(limit int) ([]models.ImageRecord, error) {
	var records []models.ImageRecord
	err := m.db.Order("processed_at DESC").Limit(limit).Find(&records).Error
	return records, err
}

// GetImageRecordByID retrieves a specific image record by ID
func (m *MetadataService) GetImageRecordByID(id uint) (*models.ImageRecord, error) {
	var record models.ImageRecord
	err := m.db.First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}
