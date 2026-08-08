package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"relaygo/relay/internal/server"
	
)

func main() {
	srv := server.New(":8080")
	


	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("server error: %v", err)
		}
	
	}()


	c := make(chan os.Signal, 1)
	//
	// 	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down relay server")
}
