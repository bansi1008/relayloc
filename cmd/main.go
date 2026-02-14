package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"relaygo/internal/server"
	"github.com/joho/godotenv"


)

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      server.NewRouter(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf(" Relay listening on :%s\n", port)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("relay crashed: %v", err)
	}
}
