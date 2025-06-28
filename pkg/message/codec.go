package message

import (
	"encoding/json"
	"time"
)

type Envelope struct {
	TraceID   string          `json:"trace_id"`
	Source    string          `json:"source"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

func Encode(traceID, source string, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	env := Envelope{
		TraceID:   traceID,
		Source:    source,
		Timestamp: time.Now().UTC(),
		Payload:   body,
	}
	return json.Marshal(env)
}

func Decode[T any](data []byte) (*Envelope, *T, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, nil, err
	}
	var payload T
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		return &env, nil, err
	}
	return &env, &payload, nil
}
