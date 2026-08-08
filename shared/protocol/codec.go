package protocol

import "encoding/json"

func Encode(frame Frame) ([]byte, error) {
	return json.Marshal(frame)
}

func Decode(data []byte) (*Frame, error) {
	var frame Frame

	if err := json.Unmarshal(data, &frame); err != nil {
		return nil, err
	}

	return &frame, nil
}
