package models

import "time"

type ImageRecord struct {
	ID          uint `gorm:"primaryKey"`
	SourceURL   string
	S3Path      string
	ProcessedAt time.Time
	Status      string // "success" / "error"
	ErrorMsg    string // nullable
	TraceID     string
	Width       int    // image width in pixels
	Height      int    // image height in pixels
	Format      string // image format (e.g., jpeg, png)
	FileSize    int64  // image file size in bytes
}

// ImageProcessedPayload represents the payload for processed image messages
type ImageProcessedPayload struct {
	SourceURL string `json:"source_url"`
	S3Path    string `json:"s3_path"`
	Status    string `json:"status"` // success/error
	ErrorMsg  string `json:"error_msg,omitempty"`
	TraceID   string `json:"trace_id"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Format    string `json:"format"`
	FileSize  int64  `json:"file_size"`
}
