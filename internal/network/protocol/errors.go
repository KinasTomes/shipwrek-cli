package protocol

import "errors"

var (
	// ErrMalformedFrame indicates that a frame is not valid JSON Lines data.
	ErrMalformedFrame = errors.New("malformed protocol frame")

	// ErrFrameTooLarge indicates that a frame exceeds MaxFrameSize.
	ErrFrameTooLarge = errors.New("protocol frame too large")

	// ErrUnsupportedVersion indicates an unsupported protocol version.
	ErrUnsupportedVersion = errors.New("unsupported protocol version")

	// ErrUnknownMessageType indicates an unknown message type.
	ErrUnknownMessageType = errors.New("unknown message type")

	// ErrInvalidEnvelope indicates invalid common envelope fields.
	ErrInvalidEnvelope = errors.New("invalid protocol envelope")

	// ErrInvalidPayload indicates that a message payload is invalid.
	ErrInvalidPayload = errors.New("invalid message payload")

	// ErrMissingMatchID indicates that a match-scoped message has no match ID.
	ErrMissingMatchID = errors.New("missing match ID")

	// ErrUnexpectedMatchID indicates that a non-match message contains a match ID.
	ErrUnexpectedMatchID = errors.New("unexpected match ID")

	// ErrInvalidSequence indicates that the sequence number is invalid.
	ErrInvalidSequence = errors.New("invalid sequence number")
)
