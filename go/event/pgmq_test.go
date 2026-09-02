package event

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestEnvelopePreservesProtobufBytes(t *testing.T) {
	payload := []byte{0x00, 0xff, 0x01, 0x80, 0x7f}
	envelope, err := NewEnvelope("command-1", "api.manage.v1.Command", payload)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := envelope.Payload()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decoded, payload) {
		t.Fatalf("payload mismatch: got %v want %v", decoded, payload)
	}
}

func TestEnvelopeRejectsInvalidContract(t *testing.T) {
	tests := []Envelope{
		{MessageType: "type", SchemaVersion: 1, CreatedAt: "2026-08-14T00:00:00Z", PayloadBase64: ""},
		{MessageID: "id", SchemaVersion: 1, CreatedAt: "2026-08-14T00:00:00Z", PayloadBase64: ""},
		{MessageID: "id", MessageType: "type", SchemaVersion: 2, CreatedAt: "2026-08-14T00:00:00Z", PayloadBase64: ""},
		{MessageID: "id", MessageType: "type", SchemaVersion: 1, CreatedAt: "not-a-time", PayloadBase64: ""},
		{MessageID: "id", MessageType: "type", SchemaVersion: 1, CreatedAt: "2026-08-14T00:00:00Z", PayloadBase64: "not base64"},
	}
	for _, envelope := range tests {
		if err := envelope.Validate(); err == nil {
			t.Fatalf("expected validation error for %+v", envelope)
		}
	}
}

func TestEnvelopePayloadUsesStandardBase64(t *testing.T) {
	envelope, err := NewEnvelope("id", "type", []byte("payload"))
	if err != nil {
		t.Fatal(err)
	}
	if envelope.PayloadBase64 != base64.StdEncoding.EncodeToString([]byte("payload")) {
		t.Fatalf("unexpected payload_base64: %s", envelope.PayloadBase64)
	}
}

func TestPGMQReadReturnsMalformedEnvelopeForArchival(t *testing.T) {
	malformed := Message{TransportID: 41}
	decodeMessageContract(&malformed, []byte(`{"message_id":`), []byte(`{}`))
	if malformed.TransportID != 41 || malformed.ContractError == "" {
		t.Fatalf("malformed row was not preserved for archival: %+v", malformed)
	}

	valid := Message{TransportID: 42}
	decodeMessageContract(&valid, mustEnvelopeJSON(t), []byte(`{}`))
	if valid.TransportID != 42 || valid.ContractError != "" {
		t.Fatalf("valid row changed: %+v", valid)
	}
}

func mustEnvelopeJSON(t *testing.T) []byte {
	t.Helper()
	envelope, err := NewEnvelope("message-1", "test.Message", []byte("payload"))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
