package main

import (
	"log"

	"relaygo/agent/internal/client"
)

func main() {
	c := client.NewClient("ws://localhost:8080/ws")

	if err := c.Connect(); err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	if err := c.Register("agent-1"); err != nil {
		log.Fatal(err)
	}

	if err := c.Read(); err != nil {
		log.Fatal(err)
	}
}
