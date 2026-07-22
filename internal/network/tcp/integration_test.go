package tcp

import (
	"context"
	"testing"
	"time"

	"github.com/KinasTomes/shipwrek-cli/internal/network/protocol"
)

func TestPeersExchangeMatchMessages(t *testing.T) {
	config := DefaultConfig()
	config.IdleTimeout = 2 * time.Second

	server, err := Listen("127.0.0.1:0", config)
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer server.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second,
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

	case <-ctx.Done():
		t.Fatal("timed out waiting for server peer")
	}

	consumeConnectedEvent(t, clientPeer)
	consumeConnectedEvent(t, serverPeer)

	const (
		matchID      = "match-integration-1"
		hostPlayerID = "host-player"
		joinPlayerID = "join-player"
	)

	// Joiner sends HELLO to host.
	clientHello := mustEnvelope(
		t,
		protocol.MessageHello,
		"",
		1,
		protocol.HelloPayload{
			PlayerID:   joinPlayerID,
			Name:       "Joiner",
			AppVersion: "0.1.0",
		},
	)

	sendEnvelope(t, clientPeer, clientHello)

	receivedClientHello := expectMessage(
		t,
		serverPeer,
		protocol.MessageHello,
	)

	clientHelloPayload, err :=
		protocol.DecodePayload[protocol.HelloPayload](
			receivedClientHello,
		)
	if err != nil {
		t.Fatalf("DecodePayload[HelloPayload]() error = %v", err)
	}

	if clientHelloPayload.PlayerID != joinPlayerID {
		t.Fatalf(
			"PlayerID = %q, want %q",
			clientHelloPayload.PlayerID,
			joinPlayerID,
		)
	}

	// Host sends HELLO to joiner.
	serverHello := mustEnvelope(
		t,
		protocol.MessageHello,
		"",
		1,
		protocol.HelloPayload{
			PlayerID:   hostPlayerID,
			Name:       "Host",
			AppVersion: "0.1.0",
		},
	)

	sendEnvelope(t, serverPeer, serverHello)

	receivedServerHello := expectMessage(
		t,
		clientPeer,
		protocol.MessageHello,
	)

	serverHelloPayload, err :=
		protocol.DecodePayload[protocol.HelloPayload](
			receivedServerHello,
		)
	if err != nil {
		t.Fatalf("DecodePayload[HelloPayload]() error = %v", err)
	}

	if serverHelloPayload.PlayerID != hostPlayerID {
		t.Fatalf(
			"PlayerID = %q, want %q",
			serverHelloPayload.PlayerID,
			hostPlayerID,
		)
	}

	// Both peers announce that placement is complete.
	clientReady := mustEnvelope(
		t,
		protocol.MessageReady,
		matchID,
		2,
		protocol.ReadyPayload{
			Ready: true,
		},
	)

	sendEnvelope(t, clientPeer, clientReady)

	receivedClientReady := expectMessage(
		t,
		serverPeer,
		protocol.MessageReady,
	)

	clientReadyPayload, err :=
		protocol.DecodePayload[protocol.ReadyPayload](
			receivedClientReady,
		)
	if err != nil {
		t.Fatalf("DecodePayload[ReadyPayload]() error = %v", err)
	}

	if !clientReadyPayload.Ready {
		t.Fatal("client Ready = false, want true")
	}

	serverReady := mustEnvelope(
		t,
		protocol.MessageReady,
		matchID,
		2,
		protocol.ReadyPayload{
			Ready: true,
		},
	)

	sendEnvelope(t, serverPeer, serverReady)

	receivedServerReady := expectMessage(
		t,
		clientPeer,
		protocol.MessageReady,
	)

	serverReadyPayload, err :=
		protocol.DecodePayload[protocol.ReadyPayload](
			receivedServerReady,
		)
	if err != nil {
		t.Fatalf("DecodePayload[ReadyPayload]() error = %v", err)
	}

	if !serverReadyPayload.Ready {
		t.Fatal("server Ready = false, want true")
	}

	// Host starts the match.
	start := mustEnvelope(
		t,
		protocol.MessageStart,
		matchID,
		3,
		protocol.StartPayload{
			FirstPlayerID: hostPlayerID,
		},
	)

	sendEnvelope(t, serverPeer, start)

	receivedStart := expectMessage(
		t,
		clientPeer,
		protocol.MessageStart,
	)

	startPayload, err :=
		protocol.DecodePayload[protocol.StartPayload](
			receivedStart,
		)
	if err != nil {
		t.Fatalf("DecodePayload[StartPayload]() error = %v", err)
	}

	if startPayload.FirstPlayerID != hostPlayerID {
		t.Fatalf(
			"FirstPlayerID = %q, want %q",
			startPayload.FirstPlayerID,
			hostPlayerID,
		)
	}

	// Host fires at D7.
	shoot := mustEnvelope(
		t,
		protocol.MessageShoot,
		matchID,
		4,
		protocol.ShootPayload{
			Turn: 1,
			X:    3,
			Y:    6,
		},
	)

	sendEnvelope(t, serverPeer, shoot)

	receivedShoot := expectMessage(
		t,
		clientPeer,
		protocol.MessageShoot,
	)

	shootPayload, err :=
		protocol.DecodePayload[protocol.ShootPayload](
			receivedShoot,
		)
	if err != nil {
		t.Fatalf("DecodePayload[ShootPayload]() error = %v", err)
	}

	if shootPayload.Turn != 1 {
		t.Fatalf("Turn = %d, want 1", shootPayload.Turn)
	}

	if shootPayload.X != 3 || shootPayload.Y != 6 {
		t.Fatalf(
			"coordinates = (%d,%d), want (3,6)",
			shootPayload.X,
			shootPayload.Y,
		)
	}

	// Joiner reports that the shot hit a ship.
	shotResult := mustEnvelope(
		t,
		protocol.MessageShotResult,
		matchID,
		3,
		protocol.ShotResultPayload{
			Turn:     1,
			X:        3,
			Y:        6,
			Result:   protocol.ShotHit,
			GameOver: false,
		},
	)

	sendEnvelope(t, clientPeer, shotResult)

	receivedShotResult := expectMessage(
		t,
		serverPeer,
		protocol.MessageShotResult,
	)

	resultPayload, err :=
		protocol.DecodePayload[protocol.ShotResultPayload](
			receivedShotResult,
		)
	if err != nil {
		t.Fatalf(
			"DecodePayload[ShotResultPayload]() error = %v",
			err,
		)
	}

	if resultPayload.Result != protocol.ShotHit {
		t.Fatalf(
			"Result = %q, want %q",
			resultPayload.Result,
			protocol.ShotHit,
		)
	}

	if resultPayload.Turn != shootPayload.Turn {
		t.Fatalf(
			"result Turn = %d, want %d",
			resultPayload.Turn,
			shootPayload.Turn,
		)
	}

	if resultPayload.X != shootPayload.X ||
		resultPayload.Y != shootPayload.Y {
		t.Fatalf(
			"result coordinates = (%d,%d), want (%d,%d)",
			resultPayload.X,
			resultPayload.Y,
			shootPayload.X,
			shootPayload.Y,
		)
	}
}

func consumeConnectedEvent(t *testing.T, peer *Peer) {
	t.Helper()

	event := waitForEvent(t, peer.Events())

	if event.Type != EventConnected {
		t.Fatalf(
			"event.Type = %q, want %q",
			event.Type,
			EventConnected,
		)
	}
}

func mustEnvelope(
	t *testing.T,
	messageType protocol.MessageType,
	matchID string,
	sequence uint64,
	payload any,
) protocol.Envelope {
	t.Helper()

	envelope, err := protocol.NewEnvelope(
		messageType,
		matchID,
		sequence,
		payload,
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	if err := protocol.ValidateEnvelope(envelope); err != nil {
		t.Fatalf("ValidateEnvelope() error = %v", err)
	}

	return envelope
}

func sendEnvelope(
	t *testing.T,
	peer *Peer,
	envelope protocol.Envelope,
) {
	t.Helper()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second,
	)
	defer cancel()

	if err := peer.Send(ctx, envelope); err != nil {
		t.Fatalf(
			"Send(%q) error = %v",
			envelope.Type,
			err,
		)
	}
}

func expectMessage(
	t *testing.T,
	peer *Peer,
	expectedType protocol.MessageType,
) protocol.Envelope {
	t.Helper()

	event := waitForEvent(t, peer.Events())

	if event.Type != EventMessage {
		t.Fatalf(
			"event.Type = %q, want %q; error = %v",
			event.Type,
			EventMessage,
			event.Err,
		)
	}

	if event.Message.Type != expectedType {
		t.Fatalf(
			"message.Type = %q, want %q",
			event.Message.Type,
			expectedType,
		)
	}

	return event.Message
}
