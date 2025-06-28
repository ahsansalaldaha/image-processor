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
}
