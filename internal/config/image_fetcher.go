package config

// ImageFetcherConfig holds configuration specific to image-fetcher service
type ImageFetcherConfig struct {
	RabbitMQ RabbitMQConfig
	Minio    MinioConfig
	Database DatabaseConfig
}

// LoadImageFetcherConfig loads configuration for image-fetcher service
func LoadImageFetcherConfig() *ImageFetcherConfig {
	return &ImageFetcherConfig{
		RabbitMQ: RabbitMQConfig{
			URL: getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
		},
		Minio: MinioConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "minio:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			UseSSL:    getEnvAsBool("MINIO_USE_SSL", false),
			Bucket:    getEnv("MINIO_BUCKET", "images"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "postgres"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "images"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
}
