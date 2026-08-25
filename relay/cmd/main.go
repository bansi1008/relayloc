package main

import (
	"context"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"relaygo/relay/internal/auth"
	"relaygo/relay/internal/db"
	"relaygo/relay/internal/server"
	"relaygo/relay/internal/user"

	"syscall"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env:", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	log.Printf("DATABASE_URL: %s", databaseURL)

	database, err := db.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	userRepo := user.NewRepository(database.Pool)

	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env:", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	tokenService := user.NewTokenService(jwtSecret)
	userService := user.NewService(userRepo, tokenService)

	authHandler := auth.NewHandler(userService)

	srv := server.New(":8080", authHandler)

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
