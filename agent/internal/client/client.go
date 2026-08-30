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
			fmt.Println("got req fram")
			var req protocol.HTTPReq

			if err := json.Unmarshal(frame.Payload, &req); err != nil {
				return err

			}
			//log.Printf("HTTP request received: %s %s", req.Method, req.Path)
			path := req.Path
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}

			url := "http://localhost:3000" + path

			if req.Query != "" {
				url += "?" + req.Query
			}

			log.Printf("urllllll %s", url)
			httpReq, err := http.NewRequest(
				req.Method,
				url,
				bytes.NewReader(req.Body),
			)

			if err != nil {
				return err
			}
			for key, values := range req.Header {
				if strings.EqualFold(key, "Accept-Encoding") {
					continue
				}

				for _, value := range values {
					httpReq.Header.Add(key, value)
				}
			}

			httpReq.Header.Set("Accept-Encoding", "identity")
			// log.Println("Agent method:", httpReq.Method)
			// log.Println("Agent headers:", httpReq.Header)
			// log.Println("Agent body:", string(req.Body))
			resp, err := http.DefaultClient.Do(httpReq)
			log.Printf("Local response status: %d", resp.StatusCode)
			log.Printf("Local response Content-Type: %q", resp.Header.Get("Content-Type"))
			log.Printf("Local response Content-Encoding: %q", resp.Header.Get("Content-Encoding"))
			log.Printf("Local response Content-Length: %q", resp.Header.Get("Content-Length"))
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			log.Printf("Body first 20 bytes: %q", body[:min(20, len(body))])
			log.Printf("Body size: %d", len(body))

			headers := make(map[string][]string)

			for key, values := range resp.Header {
				headers[key] = values
			}

			res := protocol.HTTPRes{
				ID:     req.ID,
				Status: resp.StatusCode,
				Header: headers,
				Body:   body,
			}
			payload, err := json.Marshal(res)
			if err != nil {
				return err
			}
			if err := c.WriteFrame(protocol.Frame{
				Type:    protocol.HTTPResFrame,
				Payload: payload,
			}); err != nil {
				return err
			}

		}
	}

	//return nil
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

