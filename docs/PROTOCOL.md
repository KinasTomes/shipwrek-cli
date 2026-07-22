# ShipWrek Network Protocol v1

> Application-level protocol for ShipWrek peer-to-peer LAN matches.

## 1. Scope

This document defines the protocol used after two ShipWrek clients establish a TCP connection.

It covers:

- peer introduction and version checks;
- ready-state synchronization;
- match start synchronization;
- turn-based shot exchange;
- rematch negotiation;
- heartbeat and disconnect handling;
- protocol errors.

UDP LAN discovery is specified separately.

## 2. Core rules

1. Each player keeps their fleet placement private.
2. Fleet coordinates are never transmitted during normal gameplay.
3. Each message is one UTF-8 JSON object followed by `\n`.
4. Messages are processed in sequence order.
5. Every match has a unique `match_id`.
6. Network protocol code must not depend on Bubble Tea, UI, or game-board implementations.

There is deliberately no `place` message.

## 3. Transport and framing

ShipWrek v1 uses TCP and JSON Lines.

Example wire frame:

```json
{
	"version": 1,
	"type": "shoot",
	"match_id": "m_01JABCXYZ",
	"seq": 7,
	"payload": { "turn": 4, "x": 3, "y": 6 }
}
```

The frame ends with a newline byte.

Rules:

- maximum frame size: 8192 bytes, excluding the newline;
- empty lines are invalid;
- one frame must contain exactly one JSON object;
- unknown fields are rejected;
- malformed JSON, invalid UTF-8, and oversized frames are rejected.

## 4. Coordinates

Coordinates are zero-based:

- `x`: `0..9`, representing columns `A..J`;
- `y`: `0..9`, representing rows `1..10`.

| Display | Wire            |
| ------- | --------------- |
| `A1`    | `{"x":0,"y":0}` |
| `J10`   | `{"x":9,"y":9}` |

## 5. Envelope

Every message uses:

```json
{
	"version": 1,
	"type": "shoot",
	"match_id": "m_01JABCXYZ",
	"seq": 7,
	"payload": {}
}
```

| Field      |    Type |  Required   | Description                                  |
| ---------- | ------: | :---------: | -------------------------------------------- |
| `version`  | integer |     yes     | Must be `1`.                                 |
| `type`     |  string |     yes     | Message type.                                |
| `match_id` |  string | conditional | Required for match-scoped messages.          |
| `seq`      | integer |     yes     | Per-sender sequence number, starting at `1`. |
| `payload`  |  object |     yes     | Type-specific data.                          |

### Sequence rules

Each peer maintains its own outgoing sequence counter.

- First outgoing message uses `seq = 1`.
- Each later message increments it by exactly one.
- Duplicate, lower, or skipped values are protocol errors.
- The counter continues across rematches while the TCP connection stays open.

## 6. Message types

| Type               | Direction            | Match-scoped |
| ------------------ | -------------------- | :----------: |
| `hello`            | both                 |      no      |
| `ready`            | both                 |     yes      |
| `start`            | host to joiner       |     yes      |
| `shoot`            | attacker to defender |     yes      |
| `shot_result`      | defender to attacker |     yes      |
| `ping`             | both                 |      no      |
| `pong`             | both                 |      no      |
| `rematch_request`  | both                 |     yes      |
| `rematch_response` | both                 |     yes      |
| `disconnect`       | both                 |      no      |
| `error`            | both                 | conditional  |

## 7. Messages

### 7.1 `hello`

Sent immediately after TCP connection.

```json
{
	"version": 1,
	"type": "hello",
	"seq": 1,
	"payload": {
		"player_id": "p_01JABC123",
		"name": "Rosemary",
		"app_version": "0.1.0"
	}
}
```

Payload rules:

- `player_id`: 1–64 ASCII characters;
- `name`: 1–24 Unicode code points, no control characters;
- `app_version`: 1–32 ASCII characters.

Both peers must complete `hello` before match-scoped messages are accepted.

### 7.2 `ready`

Indicates whether local fleet placement is complete.

```json
{
	"version": 1,
	"type": "ready",
	"match_id": "m_01JABCXYZ",
	"seq": 2,
	"payload": {
		"ready": true
	}
}
```

Fleet coordinates must not be included.

### 7.3 `start`

Sent by the host after both peers are ready.

```json
{
	"version": 1,
	"type": "start",
	"match_id": "m_01JABCXYZ",
	"seq": 3,
	"payload": {
		"first_player_id": "p_01JABC123"
	}
}
```

Rules:

- only the host sends `start`;
- `first_player_id` must match a player from `hello`;
- every rematch receives a new `match_id`.

### 7.4 `shoot`

Sent by the current attacker.

```json
{
	"version": 1,
	"type": "shoot",
	"match_id": "m_01JABCXYZ",
	"seq": 4,
	"payload": {
		"turn": 1,
		"x": 3,
		"y": 6
	}
}
```

Reject the message when:

- it is not the sender's turn;
- `turn` is unexpected;
- coordinates are outside `0..9`;
- the coordinate was already targeted;
- another shot is waiting for a result.

### 7.5 `shot_result`

Sent in response to one accepted `shoot`.

```json
{
	"version": 1,
	"type": "shot_result",
	"match_id": "m_01JABCXYZ",
	"seq": 5,
	"payload": {
		"turn": 1,
		"x": 3,
		"y": 6,
		"result": "hit",
		"game_over": false
	}
}
```

Payload:

| Field       | Type    | Rules                                 |
| ----------- | ------- | ------------------------------------- |
| `turn`      | integer | Must match `shoot`.                   |
| `x`         | integer | Must match `shoot`.                   |
| `y`         | integer | Must match `shoot`.                   |
| `result`    | string  | `miss`, `hit`, or `sunk`.             |
| `sunk_ship` | string  | Present only when `result` is `sunk`. |
| `game_over` | boolean | True when the final ship is sunk.     |

When `game_over` is true, `result` must be `sunk`.

The attacker must wait for `shot_result` before sending another shot.

### 7.6 `ping`

```json
{
	"version": 1,
	"type": "ping",
	"seq": 8,
	"payload": {
		"nonce": 912345678
	}
}
```

### 7.7 `pong`

```json
{
	"version": 1,
	"type": "pong",
	"seq": 9,
	"payload": {
		"nonce": 912345678
	}
}
```

The nonce must match the corresponding `ping`.

### 7.8 `rematch_request`

Valid only after game over.

```json
{
	"version": 1,
	"type": "rematch_request",
	"match_id": "m_01JABCXYZ",
	"seq": 10,
	"payload": {}
}
```

### 7.9 `rematch_response`

```json
{
	"version": 1,
	"type": "rematch_response",
	"match_id": "m_01JABCXYZ",
	"seq": 11,
	"payload": {
		"accepted": true
	}
}
```

When accepted, both players return to placement. The host later sends `start` with a new `match_id`.

### 7.10 `disconnect`

Best-effort notification before closing.

```json
{
	"version": 1,
	"type": "disconnect",
	"seq": 12,
	"payload": {
		"reason": "quit"
	}
}
```

EOF, timeout, and socket errors must also be handled as disconnects.

### 7.11 `error`

```json
{
	"version": 1,
	"type": "error",
	"match_id": "m_01JABCXYZ",
	"seq": 13,
	"payload": {
		"code": "unexpected_turn",
		"message": "expected turn 4, received turn 3",
		"fatal": false
	}
}
```

Defined codes:

| Code                   | Meaning                               |
| ---------------------- | ------------------------------------- |
| `malformed_frame`      | Invalid JSON, UTF-8, or framing.      |
| `frame_too_large`      | Frame exceeds 8192 bytes.             |
| `unsupported_version`  | Unsupported protocol version.         |
| `unknown_message_type` | Unknown `type`.                       |
| `invalid_payload`      | Invalid payload fields or values.     |
| `unexpected_message`   | Invalid for current state.            |
| `sequence_error`       | Invalid sequence number.              |
| `match_mismatch`       | Wrong `match_id`.                     |
| `unexpected_turn`      | Wrong turn number or owner.           |
| `duplicate_shot`       | Coordinate already targeted.          |
| `internal_error`       | Valid request could not be processed. |

## 8. Required flow

### Connection and start

```text
Host                                      Joiner
  |---------------- HELLO ------------------->|
  |<--------------- HELLO --------------------|
  |---------------- READY ------------------->|
  |<--------------- READY --------------------|
  |---------------- START ------------------->|
```

The host must not send `start` until both peers are ready.

### Turn exchange

```text
Attacker                                  Defender
   |---------------- SHOOT ----------------->|
   |<------------ SHOT_RESULT ---------------|
```

Only one unresolved shot may exist.

After a non-terminal result, the defender becomes the next attacker.

### Rematch

```text
Player A                                  Player B
   |----------- REMATCH_REQUEST ------------>|
   |<---------- REMATCH_RESPONSE ------------|
   |---------------- READY ----------------->|
   |<--------------- READY ------------------|
Host:
   |---------------- START ----------------->|
```

The new match must use a new `match_id`.

## 9. State validation

Recommended session states:

```text
Connected
  -> HelloComplete
  -> Placement
  -> WaitingForStart
  -> MyTurn | PeerTurn
  -> WaitingForShotResult
  -> GameOver
  -> Placement
  -> Closed
```

Examples of invalid messages:

- `shoot` before `start`;
- `ready` during combat;
- `rematch_request` before game over;
- `shot_result` without a pending local shot;
- a message with an old `match_id`.

The session layer owns state validation. The codec owns framing, JSON, and field validation.

## 10. Recommended defaults

| Setting          |    Default |
| ---------------- | ---------: |
| TCP dial timeout |  5 seconds |
| Hello timeout    | 10 seconds |
| Write timeout    |  3 seconds |
| Ping interval    |  5 seconds |
| Idle timeout     | 20 seconds |
| Maximum frame    | 8192 bytes |

## 11. Trust model

Protocol v1 is intended for friendly LAN play.

It provides:

- strict schema validation;
- frame-size limits;
- match and sequence validation;
- private fleet placement.

It does not provide:

- encryption;
- peer authentication;
- authoritative anti-cheat;
- proof that shot results are honest;
- protection from a malicious LAN peer.

## 12. Package boundary

```text
internal/network/protocol
    envelope and message types
    JSON Lines codec
    field validation

internal/network/tcp
    listen and dial
    frame read/write
    deadlines
    connection lifecycle
    network events

application/session
    handshake state
    turn validation
    match lifecycle
    rematch lifecycle
```

The protocol and TCP packages must not import Bubble Tea, UI packages, or the game board implementation.
