package websocket

import (
	"fmt"
	"net/http"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "WebSocket endpoint make")
}
