package tcp

import "errors"

var (
	ErrInvalidConfig  = errors.New("invalid TCP configuration")
	ErrPeerClosed     = errors.New("peer is closed")
	ErrEventQueueFull = errors.New("peer event queue is full")
	ErrServerClosed   = errors.New("TCP server is closed")
)
