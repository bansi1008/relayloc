package client

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"

	"encoding/json"
	"relaygo/shared/protocol"
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

		log.Printf("Received: %s", message)

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
			log.Printf("HTTP request received: %s %s", req.Method, req.Path)
			res := protocol.HTTPRes{
				ID:     req.ID,
				Status: 200,
				Header: map[string]string{},
				Body:   []byte("Hello from agent"),
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
