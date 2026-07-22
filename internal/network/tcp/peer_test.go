package tcp

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/KinasTomes/shipwrek-cli/internal/network/protocol"
)

func TestNewPeerRejectsNilConnection(t *testing.T) {
	peer, err := NewPeer(nil, DefaultConfig())

	if peer != nil {
		t.Fatal("NewPeer() peer is not nil")
	}

	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf(
			"NewPeer() error = %v, want ErrInvalidConfig",
			err,
		)
	}
}

func TestNewPeerPublishesConnectedEvent(t *testing.T) {
	localConn, remoteConn := net.Pipe()
	defer remoteConn.Close()

	peer, err := NewPeer(localConn, DefaultConfig())
	if err != nil {
		t.Fatalf("NewPeer() error = %v", err)
	}
	defer peer.Close()

	event := waitForEvent(t, peer.Events())

	if event.Type != EventConnected {
		t.Fatalf(
			"event.Type = %q, want %q",
			event.Type,
			EventConnected,
		)
	}
}

func TestPeerReceivesValidMessage(t *testing.T) {
	localConn, remoteConn := net.Pipe()
	defer remoteConn.Close()

	peer, err := NewPeer(localConn, DefaultConfig())
	if err != nil {
		t.Fatalf("NewPeer() error = %v", err)
	}
	defer peer.Close()

	// Consume connected event.
	_ = waitForEvent(t, peer.Events())

	envelope, err := protocol.NewEnvelope(
		protocol.MessagePing,
		"",
		1,
		protocol.PingPayload{
			Nonce: 123,
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	line, err := protocol.EncodeLine(envelope)
	if err != nil {
		t.Fatalf("EncodeLine() error = %v", err)
	}

	go func() {
		_, _ = remoteConn.Write(line)
	}()

	event := waitForEvent(t, peer.Events())

	if event.Type != EventMessage {
		t.Fatalf(
			"event.Type = %q, want %q",
			event.Type,
			EventMessage,
		)
	}

	if event.Message.Type != protocol.MessagePing {
		t.Fatalf(
			"message.Type = %q, want %q",
			event.Message.Type,
			protocol.MessagePing,
		)
	}

	payload, err := protocol.DecodePayload[protocol.PingPayload](
		event.Message,
	)
	if err != nil {
		t.Fatalf("DecodePayload() error = %v", err)
	}

	if payload.Nonce != 123 {
		t.Fatalf("payload.Nonce = %d, want 123", payload.Nonce)
	}
}

func TestPeerSendWritesValidMessage(t *testing.T) {
	localConn, remoteConn := net.Pipe()
	defer remoteConn.Close()

	peer, err := NewPeer(localConn, DefaultConfig())
	if err != nil {
		t.Fatalf("NewPeer() error = %v", err)
	}
	defer peer.Close()

	// Consume connected event.
	_ = waitForEvent(t, peer.Events())

	envelope, err := protocol.NewEnvelope(
		protocol.MessagePing,
		"",
		1,
		protocol.PingPayload{
			Nonce: 456,
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	received := make(chan protocol.Envelope, 1)
	readErr := make(chan error, 1)

	go func() {
		buffer := make([]byte, protocol.MaxFrameSize+1)

		count, err := remoteConn.Read(buffer)
		if err != nil {
			readErr <- err
			return
		}

		decoded, err := protocol.DecodeLine(buffer[:count])
		if err != nil {
			readErr <- err
			return
		}

		received <- decoded
	}()

	if err := peer.Send(context.Background(), envelope); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	select {
	case err := <-readErr:
		t.Fatalf("remote read error = %v", err)

	case decoded := <-received:
		if decoded.Type != protocol.MessagePing {
			t.Fatalf(
				"decoded.Type = %q, want %q",
				decoded.Type,
				protocol.MessagePing,
			)
		}

	case <-time.After(time.Second):
		t.Fatal("timed out waiting for remote message")
	}
}

func TestPeerRejectsInvalidOutgoingMessage(t *testing.T) {
	localConn, remoteConn := net.Pipe()
	defer remoteConn.Close()

	peer, err := NewPeer(localConn, DefaultConfig())
	if err != nil {
		t.Fatalf("NewPeer() error = %v", err)
	}
	defer peer.Close()

	_ = waitForEvent(t, peer.Events())

	envelope, err := protocol.NewEnvelope(
		protocol.MessageShoot,
		"",
		1,
		protocol.ShootPayload{
			Turn: 1,
			X:    3,
			Y:    6,
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	err = peer.Send(context.Background(), envelope)
	if err == nil {
		t.Fatal("Send() error = nil, want an error")
	}

	if !errors.Is(err, protocol.ErrMissingMatchID) {
		t.Fatalf(
			"Send() error = %v, want ErrMissingMatchID",
			err,
		)
	}
}

func TestPeerPublishesErrorForInvalidIncomingFrame(t *testing.T) {
	localConn, remoteConn := net.Pipe()
	defer remoteConn.Close()

	peer, err := NewPeer(localConn, DefaultConfig())
	if err != nil {
		t.Fatalf("NewPeer() error = %v", err)
	}
	defer peer.Close()

	_ = waitForEvent(t, peer.Events())

	go func() {
		_, _ = remoteConn.Write([]byte("{invalid-json}\n"))
	}()

	event := waitForEvent(t, peer.Events())

	if event.Type != EventError {
		t.Fatalf(
			"event.Type = %q, want %q",
			event.Type,
			EventError,
		)
	}

	if event.Err == nil {
		t.Fatal("event.Err = nil, want an error")
	}
}

func TestPeerPublishesDisconnectedEvent(t *testing.T) {
	localConn, remoteConn := net.Pipe()

	peer, err := NewPeer(localConn, DefaultConfig())
	if err != nil {
		t.Fatalf("NewPeer() error = %v", err)
	}
	defer peer.Close()

	_ = waitForEvent(t, peer.Events())

	if err := remoteConn.Close(); err != nil {
		t.Fatalf("remoteConn.Close() error = %v", err)
	}

	event := waitForEvent(t, peer.Events())

	if event.Type != EventDisconnected {
		t.Fatalf(
			"event.Type = %q, want %q",
			event.Type,
			EventDisconnected,
		)
	}

	if event.Err == nil {
		t.Fatal("event.Err = nil, want disconnect error")
	}
}

func TestPeerCloseIsIdempotent(t *testing.T) {
	localConn, remoteConn := net.Pipe()
	defer remoteConn.Close()

	peer, err := NewPeer(localConn, DefaultConfig())
	if err != nil {
		t.Fatalf("NewPeer() error = %v", err)
	}

	if err := peer.Close(); err != nil {
		t.Fatalf("first Close() error = %v", err)
	}

	if err := peer.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
}

func TestPeerSendAfterCloseReturnsErrPeerClosed(t *testing.T) {
	localConn, remoteConn := net.Pipe()
	defer remoteConn.Close()

	peer, err := NewPeer(localConn, DefaultConfig())
	if err != nil {
		t.Fatalf("NewPeer() error = %v", err)
	}

	if err := peer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	envelope, err := protocol.NewEnvelope(
		protocol.MessagePing,
		"",
		1,
		protocol.PingPayload{
			Nonce: 123,
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	err = peer.Send(context.Background(), envelope)

	if !errors.Is(err, ErrPeerClosed) {
		t.Fatalf(
			"Send() error = %v, want ErrPeerClosed",
			err,
		)
	}
}

func TestPeerConcurrentSendDoesNotInterleaveFrames(t *testing.T) {
	localConn, remoteConn := net.Pipe()
	defer remoteConn.Close()

	peer, err := NewPeer(localConn, DefaultConfig())
	if err != nil {
		t.Fatalf("NewPeer() error = %v", err)
	}
	defer peer.Close()

	_ = waitForEvent(t, peer.Events())

	const messageCount = 20

	readResults := make(chan protocol.Envelope, messageCount)
	readErrors := make(chan error, 1)

	go func() {
		for index := 0; index < messageCount; index++ {
			line, err := readLine(remoteConn)
			if err != nil {
				readErrors <- err
				return
			}

			envelope, err := protocol.DecodeLine(line)
			if err != nil {
				readErrors <- err
				return
			}

			readResults <- envelope
		}
	}()

	var waitGroup sync.WaitGroup
	sendErrors := make(chan error, messageCount)

	for index := 0; index < messageCount; index++ {
		waitGroup.Add(1)

		go func(sequence uint64) {
			defer waitGroup.Done()

			envelope, err := protocol.NewEnvelope(
				protocol.MessagePing,
				"",
				sequence,
				protocol.PingPayload{
					Nonce: sequence,
				},
			)
			if err != nil {
				sendErrors <- err
				return
			}

			if err := peer.Send(
				context.Background(),
				envelope,
			); err != nil {
				sendErrors <- err
			}
		}(uint64(index + 1))
	}

	waitGroup.Wait()
	close(sendErrors)

	for err := range sendErrors {
		t.Fatalf("concurrent Send() error = %v", err)
	}

	received := 0

	for received < messageCount {
		select {
		case err := <-readErrors:
			t.Fatalf("remote read error = %v", err)

		case envelope := <-readResults:
			if envelope.Type != protocol.MessagePing {
				t.Fatalf(
					"envelope.Type = %q, want %q",
					envelope.Type,
					protocol.MessagePing,
				)
			}

			received++

		case <-time.After(2 * time.Second):
			t.Fatalf(
				"timed out after receiving %d/%d messages",
				received,
				messageCount,
			)
		}
	}
}

func TestPeerReadTimeoutPublishesDisconnectedEvent(t *testing.T) {
	localConn, remoteConn := net.Pipe()
	defer remoteConn.Close()

	config := DefaultConfig()
	config.IdleTimeout = 50 * time.Millisecond

	peer, err := NewPeer(localConn, config)
	if err != nil {
		t.Fatalf("NewPeer() error = %v", err)
	}
	defer peer.Close()

	_ = waitForEvent(t, peer.Events())

	event := waitForEvent(t, peer.Events())

	if event.Type != EventDisconnected {
		t.Fatalf(
			"event.Type = %q, want %q",
			event.Type,
			EventDisconnected,
		)
	}

	if event.Err == nil {
		t.Fatal("event.Err = nil, want timeout error")
	}
}

func waitForEvent(
	t *testing.T,
	events <-chan Event,
) Event {
	t.Helper()

	select {
	case event, ok := <-events:
		if !ok {
			t.Fatal("event channel closed unexpectedly")
		}

		return event

	case <-time.After(time.Second):
		t.Fatal("timed out waiting for peer event")
		return Event{}
	}
}

func readLine(reader io.Reader) ([]byte, error) {
	buffer := make([]byte, 0, 256)
	singleByte := make([]byte, 1)

	for {
		count, err := reader.Read(singleByte)
		if err != nil {
			return nil, err
		}

		if count == 0 {
			continue
		}

		buffer = append(buffer, singleByte[0])

		if singleByte[0] == '\n' {
			return buffer, nil
		}
	}
}
