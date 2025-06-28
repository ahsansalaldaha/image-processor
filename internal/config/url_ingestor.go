package config

// URLIngestorConfig holds configuration specific to url-ingestor service
type URLIngestorConfig struct {
	Server   ServerConfig
	RabbitMQ RabbitMQConfig
}

// LoadURLIngestorConfig loads configuration for url-ingestor service
func LoadURLIngestorConfig() *URLIngestorConfig {
	return &URLIngestorConfig{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		RabbitMQ: RabbitMQConfig{
			URL: getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
		},
	}
}
