package server

import (
	"fmt"
	"net/http"
)

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from relay server!")
	})

	s.mux.HandleFunc("/ws", s.wsHandler.Handle)
	s.mux.HandleFunc("/tunnel/{id}/{path...}", s.Proxy)
}
