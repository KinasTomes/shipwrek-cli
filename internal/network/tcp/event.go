package tcp

import "github.com/KinasTomes/shipwrek-cli/internal/network/protocol"

// EventType identifies what happened to a peer connection.
type EventType string

const (
	EventConnected    EventType = "connected"
	EventMessage      EventType = "message"
	EventDisconnected EventType = "disconnected"
	EventError        EventType = "error"
)

// Event represents one event produced by a Peer.
//
// Only EventMessage contains a Message.
// EventDisconnected and EventError may contain an Err.
type Event struct {
	Type    EventType
	Message protocol.Envelope
	Err     error
}

// ConnectedEvent creates an event indicating that a peer is ready.
func ConnectedEvent() Event {
	return Event{
		Type: EventConnected,
	}
}

// MessageEvent creates an event for a received protocol message.
func MessageEvent(message protocol.Envelope) Event {
	return Event{
		Type:    EventMessage,
		Message: message,
	}
}

// DisconnectedEvent creates an event indicating that the connection ended.
func DisconnectedEvent(err error) Event {
	return Event{
		Type: EventDisconnected,
		Err:  err,
	}
}

// ErrorEvent creates an event for a non-disconnect peer error.
func ErrorEvent(err error) Event {
	return Event{
		Type: EventError,
		Err:  err,
	}
}
