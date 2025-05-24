package config

import "time"

const (
	bufferSize               = 10_000_000
	storedMessagesBufferSize = 10_000
	timeToLive               = 60 * time.Minute
)

func (cfg *Config) ApplyDefaultsValues() {
	if cfg.Subscriber.BufferSize == 0 {
		cfg.Subscriber.BufferSize = bufferSize
	}
	if cfg.Messages.TimeToLive == 0 {
		cfg.Messages.TimeToLive = 60 * timeToLive
	}
	if cfg.Messages.BufferSize == 0 {
		cfg.Messages.BufferSize = storedMessagesBufferSize
	}
}
