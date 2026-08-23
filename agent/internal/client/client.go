package client

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"

	"encoding/json"
	"relaygo/shared/protocol"
	"strings"
)

type Client struct {
	relayURL string
	conn     *websocket.Conn
	tID      string
}

func NewClient(relayURL string) *Client {
	return &Client{
		relayURL: relayURL,
	}
}

func (c *Client) Connect() error {
	//	log.Printf("Connecting to %s...", c.relayURL)

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

		//log.Printf("Received: %s", message)

		frame, err := protocol.Decode(message)
		if err != nil {
			log.Printf("Failed to decode frame: %v", err)
			continue
		}

		switch frame.Type {
		case protocol.Registered:
			log.Println("Agent registered successfully")
			var id string

			if err := json.Unmarshal(frame.Payload, &id); err != nil {
				return err
			}
			c.tID = id
			log.Printf(id)

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

			prefix := "/tunnel/" + req.ID // only if ID is the tunnel ID — otherwise don't use this

			if strings.HasPrefix(path, prefix) {
				path = strings.TrimPrefix(path, prefix)
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
