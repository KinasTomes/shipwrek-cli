package tcp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/KinasTomes/shipwrek-cli/internal/network/protocol"
)

// Peer manages one established TCP connection.
//
// Peer is responsible for:
//   - reading JSON Lines frames;
//   - decoding and validating protocol messages;
//   - publishing connection events;
//   - serializing outgoing writes;
//   - closing the underlying connection safely.
type Peer struct {
	conn   net.Conn
	config Config

	events chan Event
	done   chan struct{}

	sendMu    sync.Mutex
	closeOnce sync.Once
}

// NewPeer creates a Peer from an established TCP connection.
//
// The connection may come from either:
//   - net.Listener.Accept;
//   - net.Dialer.DialContext;
//   - net.Pipe in tests.
func NewPeer(conn net.Conn, config Config) (*Peer, error) {
	if conn == nil {
		return nil, fmt.Errorf(
			"%w: connection is nil",
			ErrInvalidConfig,
		)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	peer := &Peer{
		conn:   conn,
		config: config,
		events: make(chan Event, config.EventBufferSize),
		done:   make(chan struct{}),
	}

	// Publish the connection event before starting the reader.
	peer.events <- ConnectedEvent()

	go peer.readLoop()

	return peer, nil
}

// Events returns the read-only stream of events produced by the peer.
func (peer *Peer) Events() <-chan Event {
	return peer.events
}

// Done is closed when the peer has been shut down.
func (peer *Peer) Done() <-chan struct{} {
	return peer.done
}

// LocalAddr returns the local network address.
func (peer *Peer) LocalAddr() net.Addr {
	if peer == nil || peer.conn == nil {
		return nil
	}

	return peer.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (peer *Peer) RemoteAddr() net.Addr {
	if peer == nil || peer.conn == nil {
		return nil
	}

	return peer.conn.RemoteAddr()
}

// Send validates and sends one protocol envelope.
//
// Calls to Send are serialized so concurrent goroutines cannot interleave
// JSON frames on the same TCP connection.
func (peer *Peer) Send(
	ctx context.Context,
	envelope protocol.Envelope,
) error {
	if peer == nil {
		return ErrPeerClosed
	}

	if ctx == nil {
		return fmt.Errorf("send protocol message: context is nil")
	}

	if err := protocol.ValidateEnvelope(envelope); err != nil {
		return fmt.Errorf("send protocol message: %w", err)
	}

	line, err := protocol.EncodeLine(envelope)
	if err != nil {
		return fmt.Errorf("send protocol message: %w", err)
	}

	select {
	case <-peer.done:
		return ErrPeerClosed
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	peer.sendMu.Lock()
	defer peer.sendMu.Unlock()

	// The peer may have closed while Send was waiting for the mutex.
	select {
	case <-peer.done:
		return ErrPeerClosed
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	deadline := time.Now().Add(peer.config.WriteTimeout)

	// Respect an earlier deadline provided by the caller.
	if contextDeadline, ok := ctx.Deadline(); ok &&
		contextDeadline.Before(deadline) {
		deadline = contextDeadline
	}

	if err := peer.conn.SetWriteDeadline(deadline); err != nil {
		return fmt.Errorf("set TCP write deadline: %w", err)
	}

	defer func() {
		// Clear the deadline for future writes.
		_ = peer.conn.SetWriteDeadline(time.Time{})
	}()

	if err := writeAll(peer.conn, line); err != nil {
		if errors.Is(err, net.ErrClosed) {
			return ErrPeerClosed
		}

		return fmt.Errorf("write protocol frame: %w", err)
	}

	return nil
}

// Close shuts down the peer.
//
// Calling Close more than once is safe.
func (peer *Peer) Close() error {
	if peer == nil {
		return nil
	}

	var closeErr error

	peer.closeOnce.Do(func() {
		close(peer.done)
		closeErr = peer.conn.Close()
	})

	if closeErr != nil && !errors.Is(closeErr, net.ErrClosed) {
		return fmt.Errorf("close TCP peer: %w", closeErr)
	}

	return nil
}

// readLoop continuously reads JSON Lines frames from the connection.
func (peer *Peer) readLoop() {
	defer close(peer.events)

	scanner := bufio.NewScanner(peer.conn)

	// Scanner's default maximum token size is not part of our protocol
	// contract, so explicitly enforce the protocol frame limit.
	scanner.Buffer(
		make([]byte, 1024),
		protocol.MaxFrameSize+1,
	)

	for {
		if err := peer.conn.SetReadDeadline(
			time.Now().Add(peer.config.IdleTimeout),
		); err != nil {
			// Do not publish anything when the connection was closed locally.
			select {
			case <-peer.done:
				return
			default:
			}

			peer.publish(DisconnectedEvent(
				fmt.Errorf("set TCP read deadline: %w", err),
			))
			_ = peer.Close()
			return
		}
		if !scanner.Scan() {
			err := scanner.Err()

			if err == nil {
				err = io.EOF
			}

			// A local Close already closed done, so no additional event
			// needs to be published.
			select {
			case <-peer.done:
				return
			default:
			}

			peer.publish(DisconnectedEvent(err))
			_ = peer.Close()
			return
		}

		// Scanner reuses its backing buffer. Copy the frame before the
		// next Scan call.
		frame := append([]byte(nil), scanner.Bytes()...)

		envelope, err := protocol.DecodeLine(frame)
		if err != nil {
			peer.publish(ErrorEvent(err))
			_ = peer.Close()
			return
		}

		if err := protocol.ValidateEnvelope(envelope); err != nil {
			peer.publish(ErrorEvent(err))
			_ = peer.Close()
			return
		}

		if !peer.publish(MessageEvent(envelope)) {
			return
		}
	}
}

// publish sends one event unless the peer has already been closed.
//
// The buffered channel provides limited backpressure. If the consumer stops
// reading events, Close still unblocks this method through peer.done.
func (peer *Peer) publish(event Event) bool {
	select {
	case peer.events <- event:
		return true

	case <-peer.done:
		return false
	}
}

// writeAll writes the complete frame.
//
// TCP Write may legally write fewer bytes than requested, so a single call
// must not be assumed to send the whole JSON Lines frame.
func writeAll(writer io.Writer, data []byte) error {
	for len(data) > 0 {
		written, err := writer.Write(data)

		if err != nil {
			return err
		}

		if written == 0 {
			return io.ErrNoProgress
		}

		data = data[written:]
	}

	return nil
}
