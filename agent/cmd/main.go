package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"relaygo/internal/client"
	"relaygo/internal/server"
)

func main() {
	c := client.NewClient("ws://localhost:8080/ws")
if err := c.Connect(); err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	if err := c.Send("Hello from agent"); err != nil {
		log.Fatal(err)
	}

	if err := c.Read(); err != nil {
		log.Fatal(err)
	}


	srv := server.New(":8081")

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()
  cc := make(chan os.Signal, 1)
  signal.Notify(cc, os.Interrupt, syscall.SIGTERM)
	<-cc

	log.Println("Shutting down agent server")
}

