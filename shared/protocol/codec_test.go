package protocol

import (
	"encoding/json"
	"testing"
)

func TestHTTPRequestFrame(t *testing.T) {
	req := HTTPReq{
		ID:     "req-123",
		Method: "GET",
		Path:   "/hello",
	}

	payload, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}

	frame := Frame{
		Type:    HTTPReqFrame,
		Payload: payload,
	}

	b, err := Encode(frame)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := Decode(b)
	if err != nil {
		t.Fatal(err)
	}

	if decoded.Type != HTTPReqFrame {
		t.Errorf("expected %s, got %s", HTTPReqFrame, decoded.Type)
	}

	var decodedReq HTTPReq

	if err := json.Unmarshal(decoded.Payload, &decodedReq); err != nil {
		t.Fatal(err)
	}

	if decodedReq.ID != "req-123" {
		t.Errorf("expected ID req-123, got %s", decodedReq.ID)
	}

	if decodedReq.Method != "GET" {
		t.Errorf("expected GET, got %s", decodedReq.Method)
	}

	if decodedReq.Path != "/hello" {
		t.Errorf("expected /hello, got %s", decodedReq.Path)
	}
}
