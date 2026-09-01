package client

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"relaygo/agent/internal/store"
	"relaygo/shared/protocol"
)

type Client struct {
	relayURL  string
	conn      *websocket.Conn
	tID       string
	pubKey    string
	privKey   string
	storePath string
}

func NewClient(relayURL string) *Client {
	return &Client{
		relayURL: relayURL,
	}
}

func (c *Client) SetKeys(pubKey, privKey, storePath string) {
	c.pubKey = pubKey
	c.privKey = privKey
	c.storePath = storePath
}

func (c *Client) SetChallengeCredentials(agentID, privKey string) {
	c.tID = agentID
	c.privKey = privKey
}

func (c *Client) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.relayURL, nil)
	if err != nil {
		return err
	}

	c.conn = conn

	log.Println("Connected to relay")

	return nil
}

func (c *Client) Send(frame protocol.Frame) error {
	data, err := protocol.Encode(frame)
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.TextMessage, data)
}
func (c *Client) Read() error {
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			return err
		}

		frame, err := protocol.Decode(message)
		if err != nil {
			log.Printf("Failed to decode frame: %v", err)
			continue
		}

		switch frame.Type {
		case protocol.ChallengeResFrame:
			var res protocol.ChallengeRes
			if err := json.Unmarshal(frame.Payload, &res); err != nil {
				return err
			}

			privKeyBytes, err := base64.StdEncoding.DecodeString(c.privKey)
			if err != nil || len(privKeyBytes) != ed25519.PrivateKeySize {
				return fmt.Errorf("invalid private key in client")
			}

			sig := ed25519.Sign(privKeyBytes, []byte(res.Nonce))
			sigStr := base64.StdEncoding.EncodeToString(sig)

			verifyPayload, err := json.Marshal(protocol.ChallengeVerify{
				AgentID:   c.tID,
				Signature: sigStr,
			})
			if err != nil {
				return err
			}

			if err := c.Send(protocol.Frame{
				Type:    protocol.ChallengeVerifyFrame,
				Payload: verifyPayload,
			}); err != nil {
				return err
			}

		case protocol.AuthSuccess:
			var res protocol.AuthSuccessRes
			if err := json.Unmarshal(frame.Payload, &res); err != nil {
				return err
			}
			c.tID = res.AgentID
			log.Println(" Agent authenticated successfully!")
			log.Printf("  Agent Name: %s", res.Name)
			log.Printf("  Agent ID:   %s", res.AgentID)
			log.Printf("  Tunnel URL: %s", res.TunnelURL)

			if c.pubKey != "" && c.privKey != "" {
				targetPath := c.storePath
				if targetPath == "" {
					targetPath = "credentials.json"
				}
				creds := store.Credentials{
					AgentID:    res.AgentID,
					AgentName:  res.Name,
					PublicKey:  c.pubKey,
					PrivateKey: c.privKey,
				}
				if err := store.SaveCredentials(targetPath, creds); err != nil {
					log.Printf("Failed to save credentials: %v", err)
				} else {
					log.Printf("Saved credentials to %s", targetPath)
				}
			}

		case protocol.AuthFailed:
			var res protocol.AuthFailedRes
			if err := json.Unmarshal(frame.Payload, &res); err != nil {
				return err
			}
			log.Printf("✗ Authentication failed: %s", res.Reason)
			return fmt.Errorf("auth failed: %s", res.Reason)

		case protocol.Registered:
			log.Println("Agent registered successfully")
			var id string

			if err := json.Unmarshal(frame.Payload, &id); err != nil {
				return err
			}
			c.tID = id
			log.Println(id)

		case protocol.Ping:
			log.Println("Got PING, sending PONG")

			if err := c.Send(protocol.Frame{
				Type: protocol.Pong,
			}); err != nil {
				log.Printf("Failed to send PONG: %v", err)
				return err
			}
			log.Println("PONG sent")

		case protocol.HTTPReqFrame:
			var req protocol.HTTPReq

			if err := json.Unmarshal(frame.Payload, &req); err != nil {
				continue
			}

			path := req.Path
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}

			url := "http://localhost:3000" + path

			if req.Query != "" {
				url += "?" + req.Query
			}

			httpReq, err := http.NewRequest(
				req.Method,
				url,
				bytes.NewReader(req.Body),
			)

			if err != nil {
				res := protocol.HTTPRes{
					ID:     req.ID,
					Status: http.StatusInternalServerError,
					Header: map[string][]string{"Content-Type": {"text/plain"}},
					Body:   []byte(fmt.Sprintf("Failed to create local request: %v", err)),
				}
				if payload, err := json.Marshal(res); err == nil {
					_ = c.WriteFrame(protocol.Frame{
						Type:    protocol.HTTPResFrame,
						Payload: payload,
					})
				}
				continue
			}

			httpReq.Host = "localhost:3000"

			for key, values := range req.Header {
				lowerKey := strings.ToLower(key)
				if lowerKey == "accept-encoding" || lowerKey == "host" || lowerKey == "connection" || lowerKey == "keep-alive" || lowerKey == "transfer-encoding" || lowerKey == "upgrade" || lowerKey == "content-length" {
					continue
				}

				for _, value := range values {
					httpReq.Header.Add(key, value)
				}
			}

			httpReq.ContentLength = int64(len(req.Body))
			if len(req.Body) > 0 {
				httpReq.Body = io.NopCloser(bytes.NewReader(req.Body))
				httpReq.GetBody = func() (io.ReadCloser, error) {
					return io.NopCloser(bytes.NewReader(req.Body)), nil
				}
			}

			httpReq.Header.Set("Accept-Encoding", "identity")

			log.Printf("Forwarding %s %s (body bytes: %d, Content-Type: %q)", req.Method, url, len(req.Body), httpReq.Header.Get("Content-Type"))

			resp, err := localHTTPClient.Do(httpReq)
			if err != nil {
				log.Printf("Local request failed: %v", err)
				res := protocol.HTTPRes{
					ID:     req.ID,
					Status: http.StatusBadGateway,
					Header: map[string][]string{"Content-Type": {"text/plain"}},
					Body:   []byte(fmt.Sprintf("Bad Gateway: %v", err)),
				}
				if payload, err := json.Marshal(res); err == nil {
					_ = c.WriteFrame(protocol.Frame{
						Type:    protocol.HTTPResFrame,
						Payload: payload,
					})
				}
				continue
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Failed to read local response body: %v", err)
				body = []byte{}
			}

			headers := make(map[string][]string)
			for key, values := range resp.Header {
				lowerKey := strings.ToLower(key)
				if lowerKey == "connection" || lowerKey == "keep-alive" || lowerKey == "transfer-encoding" {
					continue
				}

				newValues := make([]string, len(values))
				for i, v := range values {
					if lowerKey == "location" {
						v = strings.ReplaceAll(v, "http://localhost:3000", "")
						v = strings.ReplaceAll(v, "http://127.0.0.1:3000", "")
						if v == "" {
							v = "/"
						}
					}
					newValues[i] = v
				}
				headers[key] = newValues
			}

			res := protocol.HTTPRes{
				ID:     req.ID,
				Status: resp.StatusCode,
				Header: headers,
				Body:   body,
			}
			payload, err := json.Marshal(res)
			if err != nil {
				log.Printf("Failed to marshal response: %v", err)
				continue
			}
			if err := c.WriteFrame(protocol.Frame{
				Type:    protocol.HTTPResFrame,
				Payload: payload,
			}); err != nil {
				log.Printf("Failed to write response frame: %v", err)
				return err
			}

		}
	}
}

var localHTTPClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Timeout: 15 * time.Second,
}
func (c *Client) WriteFrame(frame protocol.Frame) error {
	data, err := protocol.Encode(frame)
	if err != nil {
		return err
	}

	return c.conn.WriteMessage(websocket.TextMessage, data)
}
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
func (c *Client) Register(name string) error {
	payload, err := json.Marshal(name)
	if err != nil {
		return err
	}

	return c.Send(protocol.Frame{
		Type:    protocol.RegisterAgent,
		Payload: payload,
	})
}

func (c *Client) Authenticate(email, agentName, accessID, publicKey string) error {
	req := protocol.AuthReq{
		Email:     email,
		AgentName: agentName,
		AccessID:  accessID,
		PublicKey: publicKey,
	}

	log.Printf("agentttttt from case %s", req.AgentName)
	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return c.Send(protocol.Frame{
		Type:    protocol.AuthenticateAgent,
		Payload: payload,
	})
}

func (c *Client) RequestChallenge(agentID string) error {
	c.tID = agentID
	payload, err := json.Marshal(protocol.ChallengeReq{
		AgentID: agentID,
	})
	if err != nil {
		return err
	}

	return c.Send(protocol.Frame{
		Type:    protocol.ChallengeReqFrame,
		Payload: payload,
	})
}

