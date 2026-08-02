package client

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	relayURL string
	conn     *websocket.Conn
}

func NewClient(relayURL string) *Client {
	return &Client{
		relayURL: relayURL,
	}
}

func (c *Client) Connect() error {
	log.Printf("Connecting to %s...", c.relayURL)

	conn, _, err := websocket.DefaultDialer.Dial(c.relayURL, nil)
	if err != nil {
		return err
	}

	c.conn = conn

	log.Println("Connected to relay")

	return nil
}
func (c *Client) Send(message string) error {
	return c.conn.WriteMessage(
		websocket.TextMessage,
		[]byte(message),
	)
}
func (c *Client) Read() error {
	_, message, err := c.conn.ReadMessage()
	if err != nil {
		return err
	}

	log.Printf("Received: %s", message)

	return nil
}
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}