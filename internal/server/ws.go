package server

import (
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

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}

	session := tunnel.NewSession(conn)
	log.Printf("🔌 tunnel connected: %s", id)
	s.registry.Register(id, session)

	go func() {
		err := session.ReadLoop()
		log.Printf(" tunnel disconnected: %s (%v)", id, err)
		s.registry.Unregister(id)
	}()
}
