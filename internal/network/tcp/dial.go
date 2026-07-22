package tcp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
)

// Dial connects to a ShipWrek TCP server and wraps the connection in a Peer.
func Dial(
	ctx context.Context,
	address string,
	config Config,
) (*Peer, error) {
	if ctx == nil {
		return nil, fmt.Errorf("dial TCP connection: context is nil")
	}

	if strings.TrimSpace(address) == "" {
		return nil, fmt.Errorf(
			"%w: TCP address must not be empty",
			ErrInvalidConfig,
		)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Apply the configured timeout while still respecting an earlier
	// deadline or cancellation from the caller.
	dialContext, cancel := context.WithTimeout(
		ctx,
		config.DialTimeout,
	)
	defer cancel()

	dialer := net.Dialer{
		Timeout:   config.DialTimeout,
		KeepAlive: config.IdleTimeout / 2,
	}

	connection, err := dialer.DialContext(
		dialContext,
		"tcp",
		address,
	)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return nil, fmt.Errorf(
				"dial TCP connection: %w",
				context.Canceled,
			)

		case errors.Is(err, context.DeadlineExceeded):
			return nil, fmt.Errorf(
				"dial TCP connection: %w",
				context.DeadlineExceeded,
			)

		default:
			return nil, fmt.Errorf(
				"dial TCP address %q: %w",
				address,
				err,
			)
		}
	}

	peer, err := NewPeer(connection, config)
	if err != nil {
		_ = connection.Close()

		return nil, fmt.Errorf(
			"create TCP peer: %w",
			err,
		)
	}

	return peer, nil
}
