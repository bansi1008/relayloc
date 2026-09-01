package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	"relaygo/agent/internal/client"
	"relaygo/agent/internal/store"
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "agent" {
		args = args[1:]
	}

	if len(args) == 0 {
		fmt.Println("Error: no command provided")
		printUsage()
		os.Exit(1)
	}

	if args[0] != "connect" {
		fmt.Printf("Error: unknown command %q\n", args[0])
		printUsage()
		os.Exit(1)
	}

	if len(args) == 2 {
		agentName := args[1]

		creds, err := store.GetCredentialsByName("credentials.json", agentName)
		if err != nil {
			fmt.Printf("Error: agent %q is not activated on this machine.\n", agentName)
			fmt.Printf("Details: %v\n", err)
			fmt.Println("\nTo activate, run:")
			fmt.Println("  agent connect <email> <agent_name> <token>")
			os.Exit(1)
		}

		c := client.NewClient("ws://localhost:8080/ws")
		c.SetChallengeCredentials(creds.AgentID, creds.PrivateKey)

		if err := c.Connect(); err != nil {
			fmt.Printf("Error: unable to connect to relay server: %v\n", err)
			os.Exit(1)
		}
		defer c.Close()

		if err := c.RequestChallenge(creds.AgentID); err != nil {
			fmt.Printf("Error: failed to send challenge request: %v\n", err)
			os.Exit(1)
		}

		if err := c.Read(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if len(args) == 4 {
		email := args[1]
		agentName := args[2]
		token := args[3]

		pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			fmt.Printf("Error: failed to generate key pair: %v\n", err)
			os.Exit(1)
		}

		pubKeyStr := base64.StdEncoding.EncodeToString(pubKey)
		privKeyStr := base64.StdEncoding.EncodeToString(privKey)

		c := client.NewClient("ws://localhost:8080/ws")
		c.SetKeys(pubKeyStr, privKeyStr, "credentials.json")

		if err := c.Connect(); err != nil {
			fmt.Printf("Error: unable to connect to relay server: %v\n", err)
			os.Exit(1)
		}
		defer c.Close()

		if err := c.Authenticate(email, agentName, token, pubKeyStr); err != nil {
			fmt.Printf("Error: failed to send authentication request: %v\n", err)
			os.Exit(1)
		}

		if err := c.Read(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Println("Error: invalid number of arguments for 'connect'")
	printUsage()
	os.Exit(1)
}

func printUsage() {
	fmt.Println("\nUsage:")
	fmt.Println("  First-time activation:  agent connect <email> <agent_name> <token>")
	fmt.Println("  Connect active agent:    agent connect <agent_name>")
}


