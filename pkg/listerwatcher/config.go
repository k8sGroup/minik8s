package listerwatcher

import (
	"minik8s/pkg/messaging"
	"time"
)

type Config struct {
	Host        string
	HttpPort    int
	QueueConfig *messaging.QConfig
}

func DefaultConfig() *Config {
	return &Config{
		Host:        "localhost",
		HttpPort:    8080,
		QueueConfig: messaging.DefaultQConfig(),
	}
}
func GetLsConfig(host string) *Config {
	return &Config{
		Host:     host,
		HttpPort: 8080,
		QueueConfig: &messaging.QConfig{
			User:          "root",
			Password:      "123456",
			Host:          host,
			Port:          "5672",
			VHost:         "admin",
			MaxRetry:      10,
			RetryInterval: 5 * time.Second,
		},
	}
}
