package main

import (
	"encoding/json"
	"log"

	"relaygo/agent/internal/client"
	"relaygo/shared/protocol"
)

func main() {
	c := client.NewClient("ws://localhost:8080/ws")

	if err := c.Connect(); err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	payload, err := json.Marshal("agent-1")
	if err != nil {
		log.Fatal(err)
	}

	frame := protocol.Frame{
		Type:    protocol.RegisterAgent,
		Payload: payload,
	}

	if err := c.Send(frame); err != nil {
		log.Fatal(err)
	}

	if err := c.Read(); err != nil {
		log.Fatal(err)
	}
}
