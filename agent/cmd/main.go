package main

import (
	"fmt"
	"log"
	"os"

	"relaygo/agent/internal/client"
)

func main() {
	if len(os.Args) < 4 || os.Args[1] != "connect" {
		fmt.Println("Usage: agent connect <agent_name> <token>")
		os.Exit(1)
	}

	agentName := os.Args[2]
	token := os.Args[3]

	c := client.NewClient("ws://localhost:8080/ws")

	if err := c.Connect(); err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	if err := c.Authenticate(agentName, token); err != nil {
		log.Fatal(err)
	}

	if err := c.Read(); err != nil {
		log.Fatal(err)
	}
}

