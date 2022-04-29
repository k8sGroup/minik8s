package listerwatcher

import (
	"minik8s/pkg/messaging"
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
