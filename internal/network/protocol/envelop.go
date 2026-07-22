package protocol

import (
	"encoding/json"
	"fmt"
)

const (
	// CurrentVersion is the protocol version supported by this package.
	CurrentVersion uint16 = 1
)

// Envelope is the common wrapper used by every protocol message.
type Envelope struct {
	Version uint16          `json:"version"`
	Type    MessageType     `json:"type"`
	MatchID string          `json:"match_id,omitempty"`
	Seq     uint64          `json:"seq"`
	Payload json.RawMessage `json:"payload"`
}

// NewEnvelope creates an Envelope from a typed payload.
//
// The payload is encoded into JSON here, but complete validation will be
// handled separately in validation.go.
func NewEnvelope(
	messageType MessageType,
	matchID string,
	seq uint64,
	payload any,
) (Envelope, error) {
	if payload == nil {
		payload = struct{}{}
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, fmt.Errorf("encode payload: %w", err)
	}

	return Envelope{
		Version: CurrentVersion,
		Type:    messageType,
		MatchID: matchID,
		Seq:     seq,
		Payload: rawPayload,
	}, nil
}

// DecodePayload decodes an envelope payload into a concrete payload type.
//
// Example:
//
//	shoot, err := DecodePayload[ShootPayload](envelope)
func DecodePayload[T any](envelope Envelope) (T, error) {
	var payload T

	if len(envelope.Payload) == 0 {
		return payload, fmt.Errorf("decode payload: empty payload")
	}

	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return payload, fmt.Errorf(
			"decode %q payload: %w",
			envelope.Type,
			err,
		)
	}

	return payload, nil
}
