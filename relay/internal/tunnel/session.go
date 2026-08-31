package tunnel

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"relaygo/relay/internal/agent"
	"relaygo/shared/protocol"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Session struct {
	conn             *websocket.Conn
	writeMu          sync.Mutex
	respMu           sync.Mutex
	registry         *Registry
	agentService     *agent.Service
	responses        map[string]chan []byte
	id               string
	nonce            string
	challengeAgentID string
}

func NewSession(conn *websocket.Conn, registry *Registry, agentService *agent.Service) *Session {
	return &Session{
		conn:         conn,
		registry:     registry,
		agentService: agentService,
		responses:    make(map[string]chan []byte),
	}
}

func (s *Session) Register(id string) chan []byte {
	ch := make(chan []byte, 1)
	s.respMu.Lock()
	s.responses[id] = ch
	s.respMu.Unlock()
	return ch
}

func (s *Session) ReadLoop() error {
	for {
		_, msg, err := s.conn.ReadMessage()
		if err != nil {
			if s.id != "" {
				s.registry.Unregister(s.id)
				fmt.Println("it got unregistred", s.id)
			}
			return err
		}
		frame, err := protocol.Decode(msg)
		if err != nil {
			fmt.Printf("DECODE ERROR: %v\n", err)

			continue
		}
		fmt.Printf("FRAME TYPE: %s\n", frame.Type)

		switch frame.Type {
		case protocol.ChallengeReqFrame:
			var req protocol.ChallengeReq
			if err := json.Unmarshal(frame.Payload, &req); err != nil {
				log.Printf("Failed to unmarshal ChallengeReq: %v", err)
				_ = s.sendAuthFailed("invalid challenge request payload")
				continue
			}

			_, nonce, err := s.agentService.CreateChallenge(context.Background(), req.AgentID)
			if err != nil {
				log.Printf("Challenge failed for agent %s: %v", req.AgentID, err)
				_ = s.sendAuthFailed(err.Error())
				continue
			}

			s.nonce = nonce
			s.challengeAgentID = req.AgentID

			payload, err := json.Marshal(protocol.ChallengeRes{
				Nonce: nonce,
			})
			if err != nil {
				return err
			}

			if err := s.WriteFrame(protocol.Frame{
				Type:    protocol.ChallengeResFrame,
				Payload: payload,
			}); err != nil {
				return err
			}

		case protocol.ChallengeVerifyFrame:
			var req protocol.ChallengeVerify
			if err := json.Unmarshal(frame.Payload, &req); err != nil {
				log.Printf("Failed to unmarshal ChallengeVerify: %v", err)
				_ = s.sendAuthFailed("invalid challenge verify payload")
				continue
			}

			if s.nonce == "" || s.challengeAgentID != req.AgentID {
				_ = s.sendAuthFailed("invalid challenge session")
				continue
			}

			a, err := s.agentService.VerifyChallenge(context.Background(), req.AgentID, s.nonce, req.Signature)
			if err != nil {
				log.Printf("Signature verification failed for agent %s: %v", req.AgentID, err)
				_ = s.sendAuthFailed(err.Error())
				return fmt.Errorf("authentication failed: %w", err)
			}

			s.nonce = ""
			s.challengeAgentID = ""

			tunnelID := a.ID.String()
			s.registry.Register(tunnelID, s)
			s.id = tunnelID

			tunnelURL := fmt.Sprintf("http://%s.localhost:8080/", tunnelID)
			successPayload, _ := json.Marshal(protocol.AuthSuccessRes{
				AgentID:   tunnelID,
				Name:      a.Name,
				TunnelURL: tunnelURL,
			})

			if err := s.WriteFrame(protocol.Frame{
				Type:    protocol.AuthSuccess,
				Payload: successPayload,
			}); err != nil {
				log.Printf("Error sending auth success: %v", err)
				return err
			}

			log.Printf("Agent authenticated successfully: %s (id: %s)", a.Name, tunnelID)

		case protocol.AuthenticateAgent:
			var req protocol.AuthReq
			if err := json.Unmarshal(frame.Payload, &req); err != nil {
				log.Printf("Failed to unmarshal AuthReq: %v", err)
				_ = s.sendAuthFailed("invalid auth request payload")
				continue
			}

			log.Printf("Authenticating agent: email=%s, name=%s", req.Email, req.AgentName)

			a, err := s.agentService.AuthenticateWithKey(context.Background(), req.Email, req.AgentName, req.AccessID, req.PublicKey)
			if err != nil {
				log.Printf("Authentication failed for agent %s: %v", req.AgentName, err)
				_ = s.sendAuthFailed(err.Error())
				return fmt.Errorf("authentication failed: %w", err)
			}

			tunnelID := a.ID.String()
			s.registry.Register(tunnelID, s)
			s.id = tunnelID

			tunnelURL := fmt.Sprintf("http://%s.localhost:8080/", tunnelID)
			successPayload, _ := json.Marshal(protocol.AuthSuccessRes{
				AgentID:   tunnelID,
				Name:      a.Name,
				TunnelURL: tunnelURL,
			})

			if err := s.WriteFrame(protocol.Frame{
				Type:    protocol.AuthSuccess,
				Payload: successPayload,
			}); err != nil {
				log.Printf("Error sending auth success: %v", err)
				return err
			}

			log.Printf("Agent authenticated successfully: %s (id: %s)", a.Name, tunnelID)

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

			fmt.Printf("Agent registered: %s\n", id)
			payload, err := json.Marshal(id)

			if err := s.WriteFrame(protocol.Frame{
				Type:    protocol.Registered,
				Payload: payload,
			}); err != nil {
				log.Printf("Error writing ping: %v", err)
				return err
			}

		case protocol.Pong:
			fmt.Println("got pong send test")
			session, ok := s.registry.Get(s.id)
			if !ok {
				log.Print("id not found")
				continue
			}

			req := protocol.HTTPReq{
				ID:     "req-1-test",
				Method: "Get",
				Path:   "/test",
			}
			payload, err := json.Marshal(req)
			if err != nil {
				return err
			}
			if err := session.WriteFrame(protocol.Frame{
				Type:    protocol.HTTPReqFrame,
				Payload: payload,
			}); err != nil {
				return err
			}
		case protocol.HTTPResFrame:
			log.Println("got res from agent")

			var res protocol.HTTPRes
			if err := json.Unmarshal(frame.Payload, &res); err != nil {
				return err
			}
			//	log.Printf("HTTP res received: %d %s", res.Status, res.Body)
			s.Resolve(res.ID, frame.Payload)
		}

	}
}

func (s *Session) sendAuthFailed(reason string) error {
	failPayload, _ := json.Marshal(protocol.AuthFailedRes{Reason: reason})
	return s.WriteFrame(protocol.Frame{
		Type:    protocol.AuthFailed,
		Payload: failPayload,
	})
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
		s.UnregisterReq(id)
		return nil, ctx.Err()
	}
}
func (s *Session) UnregisterReq(id string) {
	//quite important if agent dose not give res then this func will remove the req and also goota put lock to prevent concurency
	s.respMu.Lock()
	defer s.respMu.Unlock()
	delete(s.responses, id)
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
	ticker := time.NewTicker(100 * time.Minute)
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
