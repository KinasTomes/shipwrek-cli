package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"
)

const (
	// MaxFrameSize is the maximum JSON frame size, excluding the newline.
	MaxFrameSize = 8 * 1024
)

// EncodeLine encodes an Envelope into one JSON Lines frame.
//
// The returned byte slice always ends with '\n'.
func EncodeLine(envelope Envelope) ([]byte, error) {
	frame, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf(
			"%w: encode envelope: %v",
			ErrMalformedFrame,
			err,
		)
	}

	if len(frame) > MaxFrameSize {
		return nil, fmt.Errorf(
			"%w: got %d bytes, maximum is %d",
			ErrFrameTooLarge,
			len(frame),
			MaxFrameSize,
		)
	}

	line := make([]byte, len(frame)+1)
	copy(line, frame)
	line[len(frame)] = '\n'

	return line, nil
}

// DecodeLine decodes one JSON Lines frame into an Envelope.
//
// The input may include a trailing '\n' or "\r\n".
func DecodeLine(line []byte) (Envelope, error) {
	frame := removeLineEnding(line)

	if len(frame) == 0 {
		return Envelope{}, fmt.Errorf(
			"%w: empty frame",
			ErrMalformedFrame,
		)
	}

	if len(frame) > MaxFrameSize {
		return Envelope{}, fmt.Errorf(
			"%w: got %d bytes, maximum is %d",
			ErrFrameTooLarge,
			len(frame),
			MaxFrameSize,
		)
	}

	if !utf8.Valid(frame) {
		return Envelope{}, fmt.Errorf(
			"%w: frame is not valid UTF-8",
			ErrMalformedFrame,
		)
	}

	// A JSON Lines frame must stay on one physical line.
	if bytes.ContainsAny(frame, "\r\n") {
		return Envelope{}, fmt.Errorf(
			"%w: frame contains an unexpected line break",
			ErrMalformedFrame,
		)
	}

	decoder := json.NewDecoder(bytes.NewReader(frame))
	decoder.DisallowUnknownFields()

	var envelope Envelope

	if err := decoder.Decode(&envelope); err != nil {
		return Envelope{}, fmt.Errorf(
			"%w: decode envelope: %v",
			ErrMalformedFrame,
			err,
		)
	}

	// Ensure the frame contains exactly one JSON value.
	var extra any

	err := decoder.Decode(&extra)
	if err != io.EOF {
		if err == nil {
			return Envelope{}, fmt.Errorf(
				"%w: frame contains multiple JSON values",
				ErrMalformedFrame,
			)
		}

		return Envelope{}, fmt.Errorf(
			"%w: unexpected trailing data: %v",
			ErrMalformedFrame,
			err,
		)
	}

	return envelope, nil
}

// WriteLine encodes an Envelope and writes the complete JSON Lines frame.
func WriteLine(writer io.Writer, envelope Envelope) error {
	if writer == nil {
		return fmt.Errorf(
			"%w: writer is nil",
			ErrMalformedFrame,
		)
	}

	line, err := EncodeLine(envelope)
	if err != nil {
		return err
	}

	written, err := writer.Write(line)
	if err != nil {
		return fmt.Errorf("write protocol frame: %w", err)
	}

	if written != len(line) {
		return fmt.Errorf(
			"write protocol frame: %w",
			io.ErrShortWrite,
		)
	}

	return nil
}

// removeLineEnding removes one optional LF or CRLF line ending.
func removeLineEnding(line []byte) []byte {
	if len(line) == 0 {
		return line
	}

	if line[len(line)-1] == '\n' {
		line = line[:len(line)-1]

		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
	}

	return line
}
