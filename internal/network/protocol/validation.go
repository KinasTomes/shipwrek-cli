package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ValidateEnvelope validates the envelope and its type-specific payload.
func ValidateEnvelope(envelope Envelope) error {
	if envelope.Version != CurrentVersion {
		return fmt.Errorf(
			"%w: got %d, expected %d",
			ErrUnsupportedVersion,
			envelope.Version,
			CurrentVersion,
		)
	}

	if !envelope.Type.IsValid() {
		return fmt.Errorf(
			"%w: %q",
			ErrUnknownMessageType,
			envelope.Type,
		)
	}

	if envelope.Seq == 0 {
		return fmt.Errorf(
			"%w: sequence must be greater than zero",
			ErrInvalidSequence,
		)
	}

	if envelope.Type.RequiresMatchID() {
		if err := validateMatchID(envelope.MatchID); err != nil {
			return err
		}
	} else if envelope.MatchID != "" && envelope.Type != MessageError {
		return fmt.Errorf(
			"%w: message %q must not contain match_id",
			ErrUnexpectedMatchID,
			envelope.Type,
		)
	}

	if err := validatePayloadObject(envelope.Payload); err != nil {
		return err
	}

	return validatePayload(envelope)
}

// validatePayload validates the payload based on the envelope message type.
func validatePayload(envelope Envelope) error {
	switch envelope.Type {
	case MessageHello:
		payload, err := decodeStrict[HelloPayload](envelope.Payload)
		if err != nil {
			return invalidPayloadError(envelope.Type, err)
		}

		return validateHelloPayload(payload)

	case MessageReady:
		_, err := decodeStrict[ReadyPayload](envelope.Payload)
		if err != nil {
			return invalidPayloadError(envelope.Type, err)
		}

		return nil

	case MessageStart:
		payload, err := decodeStrict[StartPayload](envelope.Payload)
		if err != nil {
			return invalidPayloadError(envelope.Type, err)
		}

		if err := validatePlayerID(payload.FirstPlayerID); err != nil {
			return invalidPayloadError(envelope.Type, err)
		}

		return nil

	case MessageShoot:
		payload, err := decodeStrict[ShootPayload](envelope.Payload)
		if err != nil {
			return invalidPayloadError(envelope.Type, err)
		}

		return validateShootPayload(payload)

	case MessageShotResult:
		payload, err := decodeStrict[ShotResultPayload](envelope.Payload)
		if err != nil {
			return invalidPayloadError(envelope.Type, err)
		}

		return validateShotResultPayload(payload)

	case MessagePing:
		payload, err := decodeStrict[PingPayload](envelope.Payload)
		if err != nil {
			return invalidPayloadError(envelope.Type, err)
		}

		if payload.Nonce == 0 {
			return invalidPayloadError(
				envelope.Type,
				fmt.Errorf("nonce must be greater than zero"),
			)
		}

		return nil

	case MessagePong:
		payload, err := decodeStrict[PongPayload](envelope.Payload)
		if err != nil {
			return invalidPayloadError(envelope.Type, err)
		}

		if payload.Nonce == 0 {
			return invalidPayloadError(
				envelope.Type,
				fmt.Errorf("nonce must be greater than zero"),
			)
		}

		return nil

	case MessageRematchRequest:
		_, err := decodeStrict[RematchRequestPayload](envelope.Payload)
		if err != nil {
			return invalidPayloadError(envelope.Type, err)
		}

		return nil

	case MessageRematchResponse:
		_, err := decodeStrict[RematchResponsePayload](envelope.Payload)
		if err != nil {
			return invalidPayloadError(envelope.Type, err)
		}

		return nil

	case MessageDisconnect:
		payload, err := decodeStrict[DisconnectPayload](envelope.Payload)
		if err != nil {
			return invalidPayloadError(envelope.Type, err)
		}

		if err := validateOptionalText(payload.Reason, 128); err != nil {
			return invalidPayloadError(envelope.Type, err)
		}

		return nil

	case MessageError:
		payload, err := decodeStrict[ErrorPayload](envelope.Payload)
		if err != nil {
			return invalidPayloadError(envelope.Type, err)
		}

		return validateErrorPayload(payload)

	default:
		return fmt.Errorf(
			"%w: %q",
			ErrUnknownMessageType,
			envelope.Type,
		)
	}
}

func validateHelloPayload(payload HelloPayload) error {
	if err := validatePlayerID(payload.PlayerID); err != nil {
		return invalidPayloadError(MessageHello, err)
	}

	if err := validateRequiredText(payload.Name, 24); err != nil {
		return invalidPayloadError(
			MessageHello,
			fmt.Errorf("invalid name: %w", err),
		)
	}

	if err := validateASCII(payload.AppVersion, 1, 32); err != nil {
		return invalidPayloadError(
			MessageHello,
			fmt.Errorf("invalid app_version: %w", err),
		)
	}

	return nil
}

func validateShootPayload(payload ShootPayload) error {
	if payload.Turn == 0 {
		return invalidPayloadError(
			MessageShoot,
			fmt.Errorf("turn must be greater than zero"),
		)
	}

	if payload.X > 9 {
		return invalidPayloadError(
			MessageShoot,
			fmt.Errorf("x must be between 0 and 9"),
		)
	}

	if payload.Y > 9 {
		return invalidPayloadError(
			MessageShoot,
			fmt.Errorf("y must be between 0 and 9"),
		)
	}

	return nil
}

func validateShotResultPayload(payload ShotResultPayload) error {
	if payload.Turn == 0 {
		return invalidPayloadError(
			MessageShotResult,
			fmt.Errorf("turn must be greater than zero"),
		)
	}

	if payload.X > 9 {
		return invalidPayloadError(
			MessageShotResult,
			fmt.Errorf("x must be between 0 and 9"),
		)
	}

	if payload.Y > 9 {
		return invalidPayloadError(
			MessageShotResult,
			fmt.Errorf("y must be between 0 and 9"),
		)
	}

	if !payload.Result.IsValid() {
		return invalidPayloadError(
			MessageShotResult,
			fmt.Errorf("invalid shot result %q", payload.Result),
		)
	}

	if payload.Result == ShotSunk {
		if err := validateRequiredText(payload.SunkShip, 32); err != nil {
			return invalidPayloadError(
				MessageShotResult,
				fmt.Errorf("invalid sunk_ship: %w", err),
			)
		}
	} else if payload.SunkShip != "" {
		return invalidPayloadError(
			MessageShotResult,
			fmt.Errorf("sunk_ship is only allowed when result is %q", ShotSunk),
		)
	}

	if payload.GameOver && payload.Result != ShotSunk {
		return invalidPayloadError(
			MessageShotResult,
			fmt.Errorf("game_over requires result %q", ShotSunk),
		)
	}

	return nil
}

func validateErrorPayload(payload ErrorPayload) error {
	if err := validateASCII(payload.Code, 1, 64); err != nil {
		return invalidPayloadError(
			MessageError,
			fmt.Errorf("invalid error code: %w", err),
		)
	}

	if err := validateRequiredText(payload.Message, 256); err != nil {
		return invalidPayloadError(
			MessageError,
			fmt.Errorf("invalid error message: %w", err),
		)
	}

	return nil
}

func validatePayloadObject(payload json.RawMessage) error {
	trimmed := bytes.TrimSpace(payload)

	if len(trimmed) == 0 {
		return fmt.Errorf(
			"%w: payload is empty",
			ErrInvalidPayload,
		)
	}

	if trimmed[0] != '{' || trimmed[len(trimmed)-1] != '}' {
		return fmt.Errorf(
			"%w: payload must be a JSON object",
			ErrInvalidPayload,
		)
	}

	return nil
}

func validateMatchID(matchID string) error {
	if err := validateASCII(matchID, 1, 64); err != nil {
		if matchID == "" {
			return fmt.Errorf(
				"%w: match-scoped message requires match_id",
				ErrMissingMatchID,
			)
		}

		return fmt.Errorf(
			"%w: invalid match_id: %v",
			ErrInvalidEnvelope,
			err,
		)
	}

	return nil
}

func validatePlayerID(playerID string) error {
	if err := validateASCII(playerID, 1, 64); err != nil {
		return fmt.Errorf("invalid player_id: %w", err)
	}

	return nil
}

func validateASCII(value string, minLength, maxLength int) error {
	length := len(value)

	if length < minLength || length > maxLength {
		return fmt.Errorf(
			"length must be between %d and %d bytes",
			minLength,
			maxLength,
		)
	}

	for _, character := range value {
		if character < 32 || character > 126 {
			return fmt.Errorf("must contain printable ASCII characters only")
		}
	}

	return nil
}

func validateRequiredText(value string, maxRunes int) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("must not be empty")
	}

	return validateOptionalText(value, maxRunes)
}

func validateOptionalText(value string, maxRunes int) error {
	if utf8.RuneCountInString(value) > maxRunes {
		return fmt.Errorf("must not exceed %d characters", maxRunes)
	}

	for _, character := range value {
		if unicode.IsControl(character) {
			return fmt.Errorf("must not contain control characters")
		}
	}

	return nil
}

func decodeStrict[T any](raw json.RawMessage) (T, error) {
	var payload T

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&payload); err != nil {
		return payload, err
	}

	var extra any

	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return payload, fmt.Errorf("payload contains multiple JSON values")
		}

		return payload, fmt.Errorf("unexpected trailing payload data: %w", err)
	}

	return payload, nil
}

func invalidPayloadError(messageType MessageType, err error) error {
	return fmt.Errorf(
		"%w for %q: %v",
		ErrInvalidPayload,
		messageType,
		err,
	)
}
