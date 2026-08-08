package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"relaygo/relay/internal/tunnel"
)

type Handler struct {
	registry *tunnel.Registry
}

func NewHandler(registry *tunnel.Registry) *Handler {
	return &Handler{
		registry: registry,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintln(w, "WebSocket endpoint make")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to upgrade connection:", err)
		return
	}
	defer conn.Close()
	log.Printf("Agent connected from %s", r.RemoteAddr)
	session := tunnel.NewSession(conn, h.registry)

	go func() {
		if err := session.InitPing(); err != nil {
			log.Printf("heartbeat stopped: %v", err)
		}
	}()
	if err := session.ReadLoop(); err != nil {
		log.Printf("Read error: %v", err)
	}

}
