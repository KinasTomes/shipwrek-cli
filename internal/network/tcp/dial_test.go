package tcp

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestDialRejectsNilContext(t *testing.T) {
	peer, err := Dial(
		nil,
		"127.0.0.1:12345",
		DefaultConfig(),
	)

	if peer != nil {
		t.Fatal("Dial() peer is not nil")
	}

	if err == nil {
		t.Fatal("Dial() error = nil, want an error")
	}

	if !strings.Contains(err.Error(), "context is nil") {
		t.Fatalf(
			"Dial() error = %v, want context is nil error",
			err,
		)
	}
}

func TestDialRejectsEmptyAddress(t *testing.T) {
	peer, err := Dial(
		context.Background(),
		"",
		DefaultConfig(),
	)

	if peer != nil {
		t.Fatal("Dial() peer is not nil")
	}

	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf(
			"Dial() error = %v, want ErrInvalidConfig",
			err,
		)
	}
}

func TestDialRejectsWhitespaceAddress(t *testing.T) {
	peer, err := Dial(
		context.Background(),
		"   ",
		DefaultConfig(),
	)

	if peer != nil {
		t.Fatal("Dial() peer is not nil")
	}

	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf(
			"Dial() error = %v, want ErrInvalidConfig",
			err,
		)
	}
}

func TestDialRejectsInvalidConfig(t *testing.T) {
	config := DefaultConfig()
	config.DialTimeout = 0

	peer, err := Dial(
		context.Background(),
		"127.0.0.1:12345",
		config,
	)

	if peer != nil {
		t.Fatal("Dial() peer is not nil")
	}

	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf(
			"Dial() error = %v, want ErrInvalidConfig",
			err,
		)
	}
}

func TestDialRespectsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	peer, err := Dial(
		ctx,
		"127.0.0.1:12345",
		DefaultConfig(),
	)

	if peer != nil {
		t.Fatal("Dial() peer is not nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Dial() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestDialRejectsInvalidAddress(t *testing.T) {
	peer, err := Dial(
		context.Background(),
		"invalid-address",
		DefaultConfig(),
	)

	if peer != nil {
		t.Fatal("Dial() peer is not nil")
	}

	if err == nil {
		t.Fatal("Dial() error = nil, want an error")
	}
}

func TestDialConnectsToServer(t *testing.T) {
	config := DefaultConfig()

	server, err := Listen(
		"127.0.0.1:0",
		config,
	)
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer server.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second,
	)
	defer cancel()

	acceptedPeers := make(chan *Peer, 1)
	acceptErrors := make(chan error, 1)

	go func() {
		peer, err := server.Accept(ctx)
		if err != nil {
			acceptErrors <- err
			return
		}

		acceptedPeers <- peer
	}()

	clientPeer, err := Dial(
		ctx,
		server.Addr().String(),
		config,
	)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer clientPeer.Close()

	clientEvent := waitForEvent(t, clientPeer.Events())

	if clientEvent.Type != EventConnected {
		t.Fatalf(
			"client event.Type = %q, want %q",
			clientEvent.Type,
			EventConnected,
		)
	}

	select {
	case err := <-acceptErrors:
		t.Fatalf("Accept() error = %v", err)

	case serverPeer := <-acceptedPeers:
		defer serverPeer.Close()

		serverEvent := waitForEvent(t, serverPeer.Events())

		if serverEvent.Type != EventConnected {
			t.Fatalf(
				"server event.Type = %q, want %q",
				serverEvent.Type,
				EventConnected,
			)
		}

	case <-time.After(time.Second):
		t.Fatal("timed out waiting for server peer")
	}
}

func TestDialFailsWhenNoServerIsListening(t *testing.T) {
	temporaryServer, err := Listen(
		"127.0.0.1:0",
		DefaultConfig(),
	)
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}

	address := temporaryServer.Addr().String()

	if err := temporaryServer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second,
	)
	defer cancel()

	peer, err := Dial(
		ctx,
		address,
		DefaultConfig(),
	)

	if peer != nil {
		t.Fatal("Dial() peer is not nil")
	}

	if err == nil {
		t.Fatal("Dial() error = nil, want connection error")
	}
}
