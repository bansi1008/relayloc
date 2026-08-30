package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"relaygo/agent/internal/client"
	"relaygo/agent/internal/store"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "connect" {
		fmt.Println("Usage:")
		fmt.Println("  First-time activation: agent connect <email> <agent_name> <token>")
		fmt.Println("  Connect active agent:   agent connect <agent_name>")
		os.Exit(1)
	}

	if len(os.Args) == 3 {
		agentName := os.Args[2]

		creds, err := store.GetCredentialsByName("credentials.json", agentName)
		if err != nil {
			log.Fatalf("Failed to load credentials for %q: %v\nPlease activate first using: agent connect <email> <agent_name> <token>", agentName, err)
		}

		c := client.NewClient("ws://localhost:8080/ws")
		c.SetChallengeCredentials(creds.AgentID, creds.PrivateKey)

		if err := c.Connect(); err != nil {
			log.Fatal(err)
		}
		defer c.Close()

		if err := c.RequestChallenge(creds.AgentID); err != nil {
			log.Fatal(err)
		}

		if err := c.Read(); err != nil {
			log.Fatal(err)
		}
		return
	}

	if len(os.Args) == 5 {
		email := os.Args[2]
		agentName := os.Args[3]
		token := os.Args[4]

		pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			log.Fatal(err)
		}

		pubKeyStr := base64.StdEncoding.EncodeToString(pubKey)
		privKeyStr := base64.StdEncoding.EncodeToString(privKey)

		c := client.NewClient("ws://localhost:8080/ws")
		c.SetKeys(pubKeyStr, privKeyStr, "credentials.json")

		if err := c.Connect(); err != nil {
			log.Fatal(err)
		}
		defer c.Close()

		if err := c.Authenticate(email, agentName, token, pubKeyStr); err != nil {
			log.Fatal(err)
		}

		if err := c.Read(); err != nil {
			log.Fatal(err)
		}
		return
	}

	fmt.Println("Usage:")
	fmt.Println("  First-time activation: agent connect <email> <agent_name> <token>")
	fmt.Println("  Connect active agent:   agent connect <agent_name>")
	os.Exit(1)
}

