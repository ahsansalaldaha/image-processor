package config

// ImageMetadataConfig holds configuration specific to image-metadata service
type ImageMetadataConfig struct {
	RabbitMQ RabbitMQConfig
	Database DatabaseConfig
}

// LoadImageMetadataConfig loads configuration for image-metadata service
func LoadImageMetadataConfig() *ImageMetadataConfig {
	return &ImageMetadataConfig{
		RabbitMQ: RabbitMQConfig{
			URL: getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
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
