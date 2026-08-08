package protocol

import "encoding/json"

type Frame struct {
	Type    FType           `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}
