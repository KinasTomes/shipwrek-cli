package tcp

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestListenUsesAvailablePort(t *testing.T) {
	server, err := Listen(
		"127.0.0.1:0",
		DefaultConfig(),
	)
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer server.Close()

	address := server.Addr()
	if address == nil {
		t.Fatal("Addr() = nil")
	}

	if address.String() == "" {
		t.Fatal("Addr().String() is empty")
	}
}

func TestServerAcceptsConnection(t *testing.T) {
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

	var serverPeer *Peer

	select {
	case err := <-acceptErrors:
		t.Fatalf("Accept() error = %v", err)

	case serverPeer = <-acceptedPeers:
		defer serverPeer.Close()

	case <-time.After(time.Second):
		t.Fatal("timed out waiting for accepted peer")
	}

	clientEvent := waitForEvent(t, clientPeer.Events())
	if clientEvent.Type != EventConnected {
		t.Fatalf(
			"client event.Type = %q, want %q",
			clientEvent.Type,
			EventConnected,
		)
	}

	serverEvent := waitForEvent(t, serverPeer.Events())
	if serverEvent.Type != EventConnected {
		t.Fatalf(
			"server event.Type = %q, want %q",
			serverEvent.Type,
			EventConnected,
		)
	}
}

func TestServerAcceptRespectsCanceledContext(t *testing.T) {
	server, err := Listen(
		"127.0.0.1:0",
		DefaultConfig(),
	)
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	peer, err := server.Accept(ctx)

	if peer != nil {
		t.Fatal("Accept() peer is not nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"Accept() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestServerAcceptRespectsDeadline(t *testing.T) {
	server, err := Listen(
		"127.0.0.1:0",
		DefaultConfig(),
	)
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer server.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		50*time.Millisecond,
	)
	defer cancel()

	peer, err := server.Accept(ctx)

	if peer != nil {
		t.Fatal("Accept() peer is not nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf(
			"Accept() error = %v, want context.DeadlineExceeded",
			err,
		)
	}
}

func TestServerCloseUnblocksAccept(t *testing.T) {
	server, err := Listen(
		"127.0.0.1:0",
		DefaultConfig(),
	)
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}

	acceptErrors := make(chan error, 1)

	go func() {
		_, err := server.Accept(context.Background())
		acceptErrors <- err
	}()

	if err := server.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	select {
	case err := <-acceptErrors:
		if !errors.Is(err, ErrServerClosed) {
			t.Fatalf(
				"Accept() error = %v, want ErrServerClosed",
				err,
			)
		}

	case <-time.After(time.Second):
		t.Fatal("timed out waiting for Accept() to stop")
	}
}

func TestServerCloseIsIdempotent(t *testing.T) {
	server, err := Listen(
		"127.0.0.1:0",
		DefaultConfig(),
	)
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}

	if err := server.Close(); err != nil {
		t.Fatalf("first Close() error = %v", err)
	}

	if err := server.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
}

func TestServerAcceptAfterCloseReturnsErrServerClosed(t *testing.T) {
	server, err := Listen(
		"127.0.0.1:0",
		DefaultConfig(),
	)
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}

	if err := server.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	peer, err := server.Accept(context.Background())

	if peer != nil {
		t.Fatal("Accept() peer is not nil")
	}

	if !errors.Is(err, ErrServerClosed) {
		t.Fatalf(
			"Accept() error = %v, want ErrServerClosed",
			err,
		)
	}
}

func TestServerDoneClosesAfterClose(t *testing.T) {
	server, err := Listen(
		"127.0.0.1:0",
		DefaultConfig(),
	)
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}

	select {
	case <-server.Done():
		t.Fatal("Done() closed before server.Close()")
	default:
	}

	if err := server.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	select {
	case <-server.Done():
		// Expected.
	case <-time.After(time.Second):
		t.Fatal("Done() was not closed")
	}
}
