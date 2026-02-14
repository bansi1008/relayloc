package server

import (
	"context"
	"log"
	"net/http"

	"relaygo/internal/tunnel"

	"nhooyr.io/websocket"
)

type WSServer struct {
	registry *tunnel.Registry
}

func NewWSServer(reg *tunnel.Registry) *WSServer {
	return &WSServer{registry: reg}
}

func (s *WSServer) HandleConnect(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, 
	})
	if err != nil {
		log.Printf("ws accept error: %v", err)
		return
	}

	log.Printf("🔌 tunnel connected: %s", id)
	s.registry.Register(id, conn)

	ctx := context.Background()

	
	for {
		_, _, err := conn.Read(ctx)
		if err != nil {
			break
		}
	}

	log.Printf(" tunnel disconnected: %s", id)
	s.registry.Unregister(id)
}
