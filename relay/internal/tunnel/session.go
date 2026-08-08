package tunnel

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"relaygo/shared/protocol"
	"sync"
	"time"
)

type Session struct {
	conn      *websocket.Conn
	writeMu   sync.Mutex
	respMu    sync.Mutex
	registry  *Registry
	responses map[string]chan []byte
	id        string
}

func NewSession(conn *websocket.Conn, registry *Registry) *Session {
	return &Session{
		conn:     conn,
		registry: registry,
	}
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
	for {
		//	fmt.Println("Waiting for next WebSocket message...")
		_, msg, err := s.conn.ReadMessage()
		if err != nil {
			if s.id != "" {
				s.registry.Unregister(s.id)
				fmt.Println("it got unregistred", s.id)
			}
			return err
		}
		fmt.Printf("RAW MESSAGE: %s\n", msg)
		frame, err := protocol.Decode(msg)
		if err != nil {
			fmt.Printf("DECODE ERROR: %v\n", err)

			continue
		}
		fmt.Printf("FRAME TYPE: %s\n", frame.Type)

		switch frame.Type {
		case protocol.RegisterAgent:
			var name string

			if err := json.Unmarshal(frame.Payload, &name); err != nil {
				continue
			}
			id, err := NewID()
			if err != nil {
				log.Fatal(err)
			}
			s.registry.Register(id, s)
			s.id = id
			//sess, ok := s.registry.Get(id)
			//log.Println("mapppp", sess, ok)

			fmt.Printf("Agent registered: %s\n", name)

			if err := s.WriteFrame(protocol.Frame{
				Type: protocol.Registered,
			}); err != nil {
				log.Printf("Error writing ping: %v", err)
				return err
			}
		case protocol.Pong:
			fmt.Println("got pong ")
		}

	}
}

func (s *Session) Request(ctx context.Context, id string, payload []byte) ([]byte, error) {
	ch := s.Register(id)

	if err := s.Write(payload); err != nil {
		return nil, err
	}

	select {
	case resp := <-ch:
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
func (s *Session) Write(data []byte) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	return s.conn.WriteMessage(websocket.TextMessage, data)
}

func (s *Session) WriteFrame(frame protocol.Frame) error {
	b, err := protocol.Encode(frame)
	if err != nil {
		return err
	}
	return s.Write(b)
}

func (s *Session) InitPing() error {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		err := s.WriteFrame(protocol.Frame{
			Type: protocol.Ping,
		})
		if err != nil {
			return err

		}
		fmt.Println("PING sent")
	}
	return nil
}

func NewID() (string, error) {
	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
