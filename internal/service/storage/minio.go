package storage

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"time"

	"image-processing-system/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioService handles MinIO operations
type MinioService struct {
	client *minio.Client
	config config.MinioConfig
}

// NewMinioService creates a new MinIO service instance
func NewMinioService(cfg config.MinioConfig) (*MinioService, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("Created MinIO bucket: %s", cfg.Bucket)
	}

	return &MinioService{
		client: client,
		config: cfg,
	}, nil
}

// UploadImage uploads an image to MinIO
func (m *MinioService) UploadImage(ctx context.Context, img image.Image) (string, error) {
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	filename := time.Now().Format("20060102150405") + ".jpg"
	_, err := m.client.PutObject(
		ctx,
		m.config.Bucket,
		filename,
		bytes.NewReader(buf.Bytes()),
		int64(buf.Len()),
		minio.PutObjectOptions{ContentType: "image/jpeg"},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	return filename, nil
}

// GetImageURL returns the full URL for an image
func (m *MinioService) GetImageURL(filename string) string {
	return fmt.Sprintf("s3://%s/%s", m.config.Bucket, filename)
}
