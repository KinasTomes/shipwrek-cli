package protocol

// MessageType identifies the purpose of a protocol message.
type MessageType string

const (
	MessageHello           MessageType = "hello"
	MessageReady           MessageType = "ready"
	MessageStart           MessageType = "start"
	MessageShoot           MessageType = "shoot"
	MessageShotResult      MessageType = "shot_result"
	MessagePing            MessageType = "ping"
	MessagePong            MessageType = "pong"
	MessageRematchRequest  MessageType = "rematch_request"
	MessageRematchResponse MessageType = "rematch_response"
	MessageDisconnect      MessageType = "disconnect"
	MessageError           MessageType = "error"
)

// IsValid reports whether the message type is supported by protocol v1.
func (t MessageType) IsValid() bool {
	switch t {
	case MessageHello,
		MessageReady,
		MessageStart,
		MessageShoot,
		MessageShotResult,
		MessagePing,
		MessagePong,
		MessageRematchRequest,
		MessageRematchResponse,
		MessageDisconnect,
		MessageError:
		return true
	default:
		return false
	}
}

// RequiresMatchID reports whether the message belongs to a specific match.
func (t MessageType) RequiresMatchID() bool {
	switch t {
	case MessageReady,
		MessageStart,
		MessageShoot,
		MessageShotResult,
		MessageRematchRequest,
		MessageRematchResponse:
		return true
	default:
		return false
	}
}

// HelloPayload is exchanged immediately after establishing a TCP connection.
type HelloPayload struct {
	PlayerID   string `json:"player_id"`
	Name       string `json:"name"`
	AppVersion string `json:"app_version"`
}

// ReadyPayload indicates whether the player's fleet placement is complete.
type ReadyPayload struct {
	Ready bool `json:"ready"`
}

// StartPayload announces the beginning of a match.
type StartPayload struct {
	FirstPlayerID string `json:"first_player_id"`
}

// ShootPayload represents one shot at the opponent's board.
type ShootPayload struct {
	Turn uint32 `json:"turn"`
	X    uint8  `json:"x"`
	Y    uint8  `json:"y"`
}

// ShotOutcome represents the result of a shot.
type ShotOutcome string

const (
	ShotMiss ShotOutcome = "miss"
	ShotHit  ShotOutcome = "hit"
	ShotSunk ShotOutcome = "sunk"
)

// IsValid reports whether the shot outcome is supported.
func (o ShotOutcome) IsValid() bool {
	switch o {
	case ShotMiss, ShotHit, ShotSunk:
		return true
	default:
		return false
	}
}

// ShotResultPayload responds to a ShootPayload.
type ShotResultPayload struct {
	Turn     uint32      `json:"turn"`
	X        uint8       `json:"x"`
	Y        uint8       `json:"y"`
	Result   ShotOutcome `json:"result"`
	SunkShip string      `json:"sunk_ship,omitempty"`
	GameOver bool        `json:"game_over"`
}

// PingPayload checks whether the peer is still responsive.
type PingPayload struct {
	Nonce uint64 `json:"nonce"`
}

// PongPayload responds to a PingPayload.
type PongPayload struct {
	Nonce uint64 `json:"nonce"`
}

// RematchRequestPayload intentionally contains no fields.
type RematchRequestPayload struct{}

// RematchResponsePayload accepts or rejects a rematch request.
type RematchResponsePayload struct {
	Accepted bool `json:"accepted"`
}

// DisconnectPayload provides an optional disconnect reason.
type DisconnectPayload struct {
	Reason string `json:"reason,omitempty"`
}

// ErrorPayload describes a protocol-level error.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Fatal   bool   `json:"fatal"`
}
