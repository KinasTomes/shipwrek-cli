package tcp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

const acceptPollInterval = 250 * time.Millisecond

// Server listens for one or more incoming ShipWrek TCP connections.
type Server struct {
	listener *net.TCPListener
	config   Config

	done      chan struct{}
	closeOnce sync.Once
}

// Listen creates a TCP server.
//
// Use ":0" to let the operating system choose an available port.
func Listen(address string, config Config) (*Server, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	tcpAddress, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, fmt.Errorf(
			"resolve TCP listen address %q: %w",
			address,
			err,
		)
	}

	listener, err := net.ListenTCP("tcp", tcpAddress)
	if err != nil {
		return nil, fmt.Errorf(
			"listen on TCP address %q: %w",
			address,
			err,
		)
	}

	return &Server{
		listener: listener,
		config:   config,
		done:     make(chan struct{}),
	}, nil
}

// Addr returns the address on which the server is listening.
func (server *Server) Addr() net.Addr {
	if server == nil || server.listener == nil {
		return nil
	}

	return server.listener.Addr()
}

// Done is closed when the server is shut down.
func (server *Server) Done() <-chan struct{} {
	if server == nil {
		closed := make(chan struct{})
		close(closed)
		return closed
	}

	return server.done
}

// Accept waits for one incoming connection and wraps it in a Peer.
//
// Context cancellation stops the current Accept call without permanently
// closing the server.
func (server *Server) Accept(ctx context.Context) (*Peer, error) {
	if server == nil || server.listener == nil {
		return nil, ErrServerClosed
	}

	if ctx == nil {
		return nil, fmt.Errorf("accept TCP connection: context is nil")
	}

	for {
		select {
		case <-server.done:
			return nil, ErrServerClosed

		case <-ctx.Done():
			return nil, ctx.Err()

		default:
		}

		deadline := time.Now().Add(acceptPollInterval)

		// Respect an earlier context deadline.
		if contextDeadline, ok := ctx.Deadline(); ok &&
			contextDeadline.Before(deadline) {
			deadline = contextDeadline
		}

		if err := server.listener.SetDeadline(deadline); err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil, ErrServerClosed
			}

			return nil, fmt.Errorf(
				"set TCP accept deadline: %w",
				err,
			)
		}

		connection, err := server.listener.AcceptTCP()
		if err == nil {
			// Clear the temporary accept deadline.
			_ = server.listener.SetDeadline(time.Time{})

			peer, peerErr := NewPeer(connection, server.config)
			if peerErr != nil {
				_ = connection.Close()

				return nil, fmt.Errorf(
					"create TCP peer: %w",
					peerErr,
				)
			}

			return peer, nil
		}

		var networkError net.Error

		if errors.As(err, &networkError) && networkError.Timeout() {
			// The short deadline exists so Accept can periodically
			// observe context cancellation and server shutdown.
			continue
		}

		if errors.Is(err, net.ErrClosed) {
			return nil, ErrServerClosed
		}

		return nil, fmt.Errorf(
			"accept TCP connection: %w",
			err,
		)
	}
}

// Close stops accepting new connections.
//
// Calling Close multiple times is safe. Existing Peer connections are not
// closed automatically because they own their own lifecycle.
func (server *Server) Close() error {
	if server == nil {
		return nil
	}

	var closeErr error

	server.closeOnce.Do(func() {
		close(server.done)
		closeErr = server.listener.Close()
	})

	if closeErr != nil && !errors.Is(closeErr, net.ErrClosed) {
		return fmt.Errorf("close TCP server: %w", closeErr)
	}

	return nil
}
