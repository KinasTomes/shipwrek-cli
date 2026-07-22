package protocol

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestEncodeLineAddsNewline(t *testing.T) {
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

	line, err := EncodeLine(envelope)
	if err != nil {
		t.Fatalf("EncodeLine() error = %v", err)
	}

	if len(line) == 0 {
		t.Fatal("EncodeLine() returned an empty frame")
	}

	if line[len(line)-1] != '\n' {
		t.Fatalf("EncodeLine() frame does not end with newline: %q", line)
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	original, err := NewEnvelope(
		MessageShoot,
		"match-123",
		4,
		ShootPayload{
			Turn: 1,
			X:    3,
			Y:    6,
		},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	line, err := EncodeLine(original)
	if err != nil {
		t.Fatalf("EncodeLine() error = %v", err)
	}

	decoded, err := DecodeLine(line)
	if err != nil {
		t.Fatalf("DecodeLine() error = %v", err)
	}

	if decoded.Version != original.Version {
		t.Errorf(
			"Version = %d, want %d",
			decoded.Version,
			original.Version,
		)
	}

	if decoded.Type != original.Type {
		t.Errorf(
			"Type = %q, want %q",
			decoded.Type,
			original.Type,
		)
	}

	if decoded.MatchID != original.MatchID {
		t.Errorf(
			"MatchID = %q, want %q",
			decoded.MatchID,
			original.MatchID,
		)
	}

	if decoded.Seq != original.Seq {
		t.Errorf(
			"Seq = %d, want %d",
			decoded.Seq,
			original.Seq,
		)
	}

	payload, err := DecodePayload[ShootPayload](decoded)
	if err != nil {
		t.Fatalf("DecodePayload() error = %v", err)
	}

	if payload.Turn != 1 {
		t.Errorf("Turn = %d, want 1", payload.Turn)
	}

	if payload.X != 3 {
		t.Errorf("X = %d, want 3", payload.X)
	}

	if payload.Y != 6 {
		t.Errorf("Y = %d, want 6", payload.Y)
	}
}

func TestDecodeLineAcceptsCRLF(t *testing.T) {
	line := []byte(
		"{\"version\":1,\"type\":\"ping\",\"seq\":1," +
			"\"payload\":{\"nonce\":123}}\r\n",
	)

	envelope, err := DecodeLine(line)
	if err != nil {
		t.Fatalf("DecodeLine() error = %v", err)
	}

	if envelope.Type != MessagePing {
		t.Errorf(
			"Type = %q, want %q",
			envelope.Type,
			MessagePing,
		)
	}
}

func TestDecodeLineRejectsEmptyFrame(t *testing.T) {
	_, err := DecodeLine([]byte("\n"))
	if err == nil {
		t.Fatal("DecodeLine() error = nil, want an error")
	}

	if !errors.Is(err, ErrMalformedFrame) {
		t.Fatalf(
			"DecodeLine() error = %v, want ErrMalformedFrame",
			err,
		)
	}
}

func TestDecodeLineRejectsMalformedJSON(t *testing.T) {
	line := []byte(
		`{"version":1,"type":"ping","seq":1,"payload":`,
	)

	_, err := DecodeLine(line)
	if err == nil {
		t.Fatal("DecodeLine() error = nil, want an error")
	}

	if !errors.Is(err, ErrMalformedFrame) {
		t.Fatalf(
			"DecodeLine() error = %v, want ErrMalformedFrame",
			err,
		)
	}
}

func TestDecodeLineRejectsUnknownEnvelopeField(t *testing.T) {
	line := []byte(
		`{"version":1,"type":"ping","seq":1,` +
			`"unexpected":"value","payload":{"nonce":123}}`,
	)

	_, err := DecodeLine(line)
	if err == nil {
		t.Fatal("DecodeLine() error = nil, want an error")
	}

	if !errors.Is(err, ErrMalformedFrame) {
		t.Fatalf(
			"DecodeLine() error = %v, want ErrMalformedFrame",
			err,
		)
	}
}

func TestDecodeLineRejectsMultipleJSONValues(t *testing.T) {
	line := []byte(
		`{"version":1,"type":"ping","seq":1,"payload":{"nonce":1}}` +
			`{"version":1,"type":"ping","seq":2,"payload":{"nonce":2}}`,
	)

	_, err := DecodeLine(line)
	if err == nil {
		t.Fatal("DecodeLine() error = nil, want an error")
	}

	if !errors.Is(err, ErrMalformedFrame) {
		t.Fatalf(
			"DecodeLine() error = %v, want ErrMalformedFrame",
			err,
		)
	}
}

func TestDecodeLineRejectsUnexpectedLineBreak(t *testing.T) {
	line := []byte(
		"{\"version\":1,\n\"type\":\"ping\"," +
			"\"seq\":1,\"payload\":{\"nonce\":123}}",
	)

	_, err := DecodeLine(line)
	if err == nil {
		t.Fatal("DecodeLine() error = nil, want an error")
	}

	if !errors.Is(err, ErrMalformedFrame) {
		t.Fatalf(
			"DecodeLine() error = %v, want ErrMalformedFrame",
			err,
		)
	}
}

func TestDecodeLineRejectsInvalidUTF8(t *testing.T) {
	line := []byte{
		'{',
		'"', 'x', '"', ':',
		'"', 0xff, '"',
		'}',
	}

	_, err := DecodeLine(line)
	if err == nil {
		t.Fatal("DecodeLine() error = nil, want an error")
	}

	if !errors.Is(err, ErrMalformedFrame) {
		t.Fatalf(
			"DecodeLine() error = %v, want ErrMalformedFrame",
			err,
		)
	}
}

func TestDecodeLineRejectsOversizedFrame(t *testing.T) {
	line := []byte(strings.Repeat("a", MaxFrameSize+1))

	_, err := DecodeLine(line)
	if err == nil {
		t.Fatal("DecodeLine() error = nil, want an error")
	}

	if !errors.Is(err, ErrFrameTooLarge) {
		t.Fatalf(
			"DecodeLine() error = %v, want ErrFrameTooLarge",
			err,
		)
	}
}

func TestWriteLineWritesCompleteFrame(t *testing.T) {
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

	var buffer bytes.Buffer

	if err := WriteLine(&buffer, envelope); err != nil {
		t.Fatalf("WriteLine() error = %v", err)
	}

	if buffer.Len() == 0 {
		t.Fatal("WriteLine() wrote no data")
	}

	if !bytes.HasSuffix(buffer.Bytes(), []byte{'\n'}) {
		t.Fatalf(
			"WriteLine() output does not end with newline: %q",
			buffer.Bytes(),
		)
	}
}

func TestWriteLineRejectsNilWriter(t *testing.T) {
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

	err = WriteLine(nil, envelope)
	if err == nil {
		t.Fatal("WriteLine() error = nil, want an error")
	}
}
