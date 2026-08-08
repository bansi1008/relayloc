package protocol

import (
	"encoding/json"
	"testing"
)

func TestEncode(t *testing.T) {
	p, err:=json.Marshal("agent-1")
	if err != nil {
		t.Error(err)
	}
	frame := Frame{
		Type: RegisterAgent,
		Payload: p,
	}

	b, err := Encode(frame)
	if err != nil {
		t.Error(err)
	}

	decoded, err := Decode(b)
	if err != nil {
		t.Error(err)
	}

	if decoded.Type != RegisterAgent {
		t.Errorf("expected type %s, got %s", RegisterAgent, decoded.Type)
	}

	var name string
	if err := json.Unmarshal(decoded.Payload, &name); err != nil {
    	t.Error(err)
	}

	if name != "agent-1" {
    	t.Errorf("expected payload name %s, got %s", "agent-1", name)
	}

}	
