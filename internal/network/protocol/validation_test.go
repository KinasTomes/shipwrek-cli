package protocol

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestValidateEnvelopeAcceptsValidShoot(t *testing.T) {
	envelope, err := NewEnvelope(
		MessageShoot,
		"match-123",
		1,
		ShootPayload{
			Turn: 1,
			X:    3,
			Y:    6,
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	if err := ValidateEnvelope(envelope); err != nil {
		t.Fatalf("ValidateEnvelope() error = %v", err)
	}
}

func TestValidateEnvelopeRejectsUnsupportedVersion(t *testing.T) {
	envelope, err := NewEnvelope(
		MessagePing,
		"",
		1,
		PingPayload{
			Nonce: 123,
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	envelope.Version = 2

	err = ValidateEnvelope(envelope)
	if err == nil {
		t.Fatal("ValidateEnvelope() error = nil, want an error")
	}

	if !errors.Is(err, ErrUnsupportedVersion) {
		t.Fatalf(
			"ValidateEnvelope() error = %v, want ErrUnsupportedVersion",
			err,
		)
	}
}

func TestValidateEnvelopeRejectsUnknownMessageType(t *testing.T) {
	envelope := Envelope{
		Version: CurrentVersion,
		Type:    MessageType("dance"),
		Seq:     1,
		Payload: json.RawMessage(`{}`),
	}

	err := ValidateEnvelope(envelope)
	if err == nil {
		t.Fatal("ValidateEnvelope() error = nil, want an error")
	}

	if !errors.Is(err, ErrUnknownMessageType) {
		t.Fatalf(
			"ValidateEnvelope() error = %v, want ErrUnknownMessageType",
			err,
		)
	}
}

func TestValidateEnvelopeRejectsZeroSequence(t *testing.T) {
	envelope, err := NewEnvelope(
		MessagePing,
		"",
		0,
		PingPayload{
			Nonce: 123,
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	err = ValidateEnvelope(envelope)
	if err == nil {
		t.Fatal("ValidateEnvelope() error = nil, want an error")
	}

	if !errors.Is(err, ErrInvalidSequence) {
		t.Fatalf(
			"ValidateEnvelope() error = %v, want ErrInvalidSequence",
			err,
		)
	}
}

func TestValidateEnvelopeRejectsMissingMatchID(t *testing.T) {
	envelope, err := NewEnvelope(
		MessageShoot,
		"",
		1,
		ShootPayload{
			Turn: 1,
			X:    3,
			Y:    6,
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	err = ValidateEnvelope(envelope)
	if err == nil {
		t.Fatal("ValidateEnvelope() error = nil, want an error")
	}

	if !errors.Is(err, ErrMissingMatchID) {
		t.Fatalf(
			"ValidateEnvelope() error = %v, want ErrMissingMatchID",
			err,
		)
	}
}

func TestValidateEnvelopeRejectsUnexpectedMatchID(t *testing.T) {
	envelope, err := NewEnvelope(
		MessagePing,
		"match-123",
		1,
		PingPayload{
			Nonce: 123,
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	err = ValidateEnvelope(envelope)
	if err == nil {
		t.Fatal("ValidateEnvelope() error = nil, want an error")
	}

	if !errors.Is(err, ErrUnexpectedMatchID) {
		t.Fatalf(
			"ValidateEnvelope() error = %v, want ErrUnexpectedMatchID",
			err,
		)
	}
}

func TestValidateEnvelopeRejectsEmptyPayload(t *testing.T) {
	envelope := Envelope{
		Version: CurrentVersion,
		Type:    MessagePing,
		Seq:     1,
		Payload: nil,
	}

	err := ValidateEnvelope(envelope)
	if err == nil {
		t.Fatal("ValidateEnvelope() error = nil, want an error")
	}

	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf(
			"ValidateEnvelope() error = %v, want ErrInvalidPayload",
			err,
		)
	}
}

func TestValidateEnvelopeRejectsNonObjectPayload(t *testing.T) {
	envelope := Envelope{
		Version: CurrentVersion,
		Type:    MessagePing,
		Seq:     1,
		Payload: json.RawMessage(`"hello"`),
	}

	err := ValidateEnvelope(envelope)
	if err == nil {
		t.Fatal("ValidateEnvelope() error = nil, want an error")
	}

	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf(
			"ValidateEnvelope() error = %v, want ErrInvalidPayload",
			err,
		)
	}
}

func TestValidateEnvelopeRejectsUnknownPayloadField(t *testing.T) {
	envelope := Envelope{
		Version: CurrentVersion,
		Type:    MessagePing,
		Seq:     1,
		Payload: json.RawMessage(
			`{"nonce":123,"unexpected":"value"}`,
		),
	}

	err := ValidateEnvelope(envelope)
	if err == nil {
		t.Fatal("ValidateEnvelope() error = nil, want an error")
	}

	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf(
			"ValidateEnvelope() error = %v, want ErrInvalidPayload",
			err,
		)
	}
}

func TestValidateShootRejectsZeroTurn(t *testing.T) {
	envelope, err := NewEnvelope(
		MessageShoot,
		"match-123",
		1,
		ShootPayload{
			Turn: 0,
			X:    3,
			Y:    6,
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	err = ValidateEnvelope(envelope)
	if err == nil {
		t.Fatal("ValidateEnvelope() error = nil, want an error")
	}

	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf(
			"ValidateEnvelope() error = %v, want ErrInvalidPayload",
			err,
		)
	}
}

func TestValidateShootRejectsCoordinateOutsideBoard(t *testing.T) {
	tests := []struct {
		name    string
		payload ShootPayload
	}{
		{
			name: "x outside board",
			payload: ShootPayload{
				Turn: 1,
				X:    10,
				Y:    5,
			},
		},
		{
			name: "y outside board",
			payload: ShootPayload{
				Turn: 1,
				X:    5,
				Y:    10,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			envelope, err := NewEnvelope(
				MessageShoot,
				"match-123",
				1,
				test.payload,
			)
			if err != nil {
				t.Fatalf("NewEnvelope() error = %v", err)
			}

			err = ValidateEnvelope(envelope)
			if err == nil {
				t.Fatal(
					"ValidateEnvelope() error = nil, want an error",
				)
			}

			if !errors.Is(err, ErrInvalidPayload) {
				t.Fatalf(
					"ValidateEnvelope() error = %v, want ErrInvalidPayload",
					err,
				)
			}
		})
	}
}

func TestValidateShotResultAcceptsSunkResult(t *testing.T) {
	envelope, err := NewEnvelope(
		MessageShotResult,
		"match-123",
		1,
		ShotResultPayload{
			Turn:     4,
			X:        3,
			Y:        6,
			Result:   ShotSunk,
			SunkShip: "destroyer",
			GameOver: true,
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	if err := ValidateEnvelope(envelope); err != nil {
		t.Fatalf("ValidateEnvelope() error = %v", err)
	}
}

func TestValidateShotResultRejectsInvalidOutcome(t *testing.T) {
	envelope, err := NewEnvelope(
		MessageShotResult,
		"match-123",
		1,
		ShotResultPayload{
			Turn:   1,
			X:      3,
			Y:      6,
			Result: ShotOutcome("destroyed"),
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	err = ValidateEnvelope(envelope)
	if err == nil {
		t.Fatal("ValidateEnvelope() error = nil, want an error")
	}

	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf(
			"ValidateEnvelope() error = %v, want ErrInvalidPayload",
			err,
		)
	}
}

func TestValidateShotResultRejectsSunkShipForHit(t *testing.T) {
	envelope, err := NewEnvelope(
		MessageShotResult,
		"match-123",
		1,
		ShotResultPayload{
			Turn:     1,
			X:        3,
			Y:        6,
			Result:   ShotHit,
			SunkShip: "destroyer",
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	err = ValidateEnvelope(envelope)
	if err == nil {
		t.Fatal("ValidateEnvelope() error = nil, want an error")
	}

	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf(
			"ValidateEnvelope() error = %v, want ErrInvalidPayload",
			err,
		)
	}
}

func TestValidateShotResultRejectsGameOverWithoutSunk(t *testing.T) {
	envelope, err := NewEnvelope(
		MessageShotResult,
		"match-123",
		1,
		ShotResultPayload{
			Turn:     1,
			X:        3,
			Y:        6,
			Result:   ShotHit,
			GameOver: true,
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	err = ValidateEnvelope(envelope)
	if err == nil {
		t.Fatal("ValidateEnvelope() error = nil, want an error")
	}

	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf(
			"ValidateEnvelope() error = %v, want ErrInvalidPayload",
			err,
		)
	}
}

func TestValidatePingRejectsZeroNonce(t *testing.T) {
	envelope, err := NewEnvelope(
		MessagePing,
		"",
		1,
		PingPayload{
			Nonce: 0,
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	err = ValidateEnvelope(envelope)
	if err == nil {
		t.Fatal("ValidateEnvelope() error = nil, want an error")
	}

	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf(
			"ValidateEnvelope() error = %v, want ErrInvalidPayload",
			err,
		)
	}
}

func TestValidateHelloAcceptsValidUnicodeName(t *testing.T) {
	envelope, err := NewEnvelope(
		MessageHello,
		"",
		1,
		HelloPayload{
			PlayerID:   "player-123",
			Name:       "Phước",
			AppVersion: "0.1.0",
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	if err := ValidateEnvelope(envelope); err != nil {
		t.Fatalf("ValidateEnvelope() error = %v", err)
	}
}

func TestValidateHelloRejectsControlCharacters(t *testing.T) {
	envelope, err := NewEnvelope(
		MessageHello,
		"",
		1,
		HelloPayload{
			PlayerID:   "player-123",
			Name:       "Player\nTwo",
			AppVersion: "0.1.0",
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	err = ValidateEnvelope(envelope)
	if err == nil {
		t.Fatal("ValidateEnvelope() error = nil, want an error")
	}

	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf(
			"ValidateEnvelope() error = %v, want ErrInvalidPayload",
			err,
		)
	}
}

func TestValidateHelloRejectsNonASCIIAppVersion(t *testing.T) {
	envelope, err := NewEnvelope(
		MessageHello,
		"",
		1,
		HelloPayload{
			PlayerID:   "player-123",
			Name:       "Player",
			AppVersion: "vẻ-1",
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	err = ValidateEnvelope(envelope)
	if err == nil {
		t.Fatal("ValidateEnvelope() error = nil, want an error")
	}

	if !errors.Is(err, ErrInvalidPayload) {
		t.Fatalf(
			"ValidateEnvelope() error = %v, want ErrInvalidPayload",
			err,
		)
	}
}

func TestValidateErrorAllowsMatchID(t *testing.T) {
	envelope, err := NewEnvelope(
		MessageError,
		"match-123",
		1,
		ErrorPayload{
			Code:    "unexpected_turn",
			Message: "expected turn 4",
			Fatal:   false,
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	if err := ValidateEnvelope(envelope); err != nil {
		t.Fatalf("ValidateEnvelope() error = %v", err)
	}
}
