package server

import (
	//	"fmt"
	//	"net/http"
	"relaygo/relay/internal/auth"
)

func (s *Server) registerRoutes(authHandler *auth.Handler) {
	// s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintln(w, "Hello from relay server!")
	// })

	s.mux.HandleFunc("/ws", s.wsHandler.Handle)
	s.mux.HandleFunc("/tunnel/{id}/{path...}", s.Proxy)
	s.mux.HandleFunc("/auth/register", authHandler.Register)
}
