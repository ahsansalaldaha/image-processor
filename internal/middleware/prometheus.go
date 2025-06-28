package middleware

import (
	"github.com/prometheus/client_golang/prometheus"
)

// WorkerMetrics holds all worker-related Prometheus metrics
var (
	// Image processing metrics
	ImagesProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "images_processed_total",
			Help: "Total number of images processed",
		},
		[]string{"status", "service"},
	)

	ProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "image_processing_duration_seconds",
			Help:    "Image processing duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"step", "service"},
	)

	// Queue metrics
	QueueSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queue_size",
			Help: "Current size of the processing queue",
		},
		[]string{"queue_name", "service"},
	)

	ActiveWorkers = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_workers",
			Help: "Number of currently active workers",
		},
		[]string{"service"},
	)

	// Job processing metrics
	JobsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_processed_total",
			Help: "Total number of jobs processed",
		},
		[]string{"status", "service"},
	)

	JobProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "job_processing_duration_seconds",
			Help:    "Job processing duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service"},
	)
)

func init() {
	// Register all worker metrics
	prometheus.MustRegister(ImagesProcessed)
	prometheus.MustRegister(ProcessingDuration)
	prometheus.MustRegister(QueueSize)
	prometheus.MustRegister(ActiveWorkers)
	prometheus.MustRegister(JobsProcessed)
	prometheus.MustRegister(JobProcessingDuration)
}
