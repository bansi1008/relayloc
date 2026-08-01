package tunnel

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

type Session struct {
	Conn      *websocket.Conn
	writeMu   sync.Mutex
	respMu    sync.Mutex
	responses map[string]chan []byte
}

func NewSession(conn *websocket.Conn) *Session {
	return &Session{
		Conn:      conn,
		responses: make(map[string]chan []byte),
	}
}

func (s *Session) Write(ctx context.Context, data []byte) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.Conn.Write(ctx, websocket.MessageText, data)
}

func (s *Session) Register(id string) chan []byte {
	ch := make(chan []byte, 1)
	s.respMu.Lock()
	s.responses[id] = ch
	s.respMu.Unlock()
	return ch
}

func (s *Session) Resolve(id string, msg []byte) {
	s.respMu.Lock()
	if ch, ok := s.responses[id]; ok {
		ch <- msg
		close(ch)
		delete(s.responses, id)
	}
	s.respMu.Unlock()
}

func (s *Session) ReadLoop() error {
	ctx := context.Background()
	for {
		_, msg, err := s.Conn.Read(ctx)
		if err != nil {
			return err
		}

		var env struct {
			Type string `json:"type"`
			ID   string `json:"id"`
		}
		if err := json.Unmarshal(msg, &env); err != nil {
			continue
		}

		if env.Type == "http_response" {
			s.Resolve(env.ID, msg)
		}
	}
}

func (s *Session) Request(ctx context.Context, id string, payload []byte) ([]byte, error) {
	ch := s.Register(id)
	if err := s.Write(ctx, payload); err != nil {
		return nil, err
	}

	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(30 * time.Second):
		return nil, errors.New("agent timeout")
	}
}
