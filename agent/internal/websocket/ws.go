package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
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
	for {
		_, msggg, err := conn.ReadMessage()
		err = conn.WriteMessage(websocket.TextMessage, []byte("Hello from relay server!"))
		log.Printf("Received message from %s: %s", string(msggg))

		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
	}

}
