package tcp

import "time"

// Config contains TCP connection and peer lifecycle settings.
type Config struct {
	DialTimeout      time.Duration
	HandshakeTimeout time.Duration
	WriteTimeout     time.Duration
	IdleTimeout      time.Duration
	EventBufferSize  int
}

// DefaultConfig returns the default TCP settings for ShipWrek.
func DefaultConfig() Config {
	return Config{
		DialTimeout:      5 * time.Second,
		HandshakeTimeout: 10 * time.Second,
		WriteTimeout:     3 * time.Second,
		IdleTimeout:      20 * time.Second,
		EventBufferSize:  32,
	}
}

// Validate checks whether the TCP configuration is usable.
func (config Config) Validate() error {
	if config.DialTimeout <= 0 {
		return ErrInvalidConfig
	}

	if config.HandshakeTimeout <= 0 {
		return ErrInvalidConfig
	}

	if config.WriteTimeout <= 0 {
		return ErrInvalidConfig
	}

	if config.IdleTimeout <= 0 {
		return ErrInvalidConfig
	}

	if config.EventBufferSize <= 0 {
		return ErrInvalidConfig
	}

	return nil
}
