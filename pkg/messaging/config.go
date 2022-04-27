package messaging

import "time"

type QConfig struct {
	User          string
	Password      string
	Host          string
	Port          string
	MaxRetry      int
	RetryInterval time.Duration
}

func DefaultQConfig() *QConfig {
	config := QConfig{
		User:          "root",
		Password:      "123456",
		Host:          "localhost",
		Port:          "5672",
		MaxRetry:      10,
		RetryInterval: 5 * time.Second,
	}
	return &config
}
