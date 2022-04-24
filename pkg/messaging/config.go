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
