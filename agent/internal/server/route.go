package server

import (
	"fmt"
	"net/http"
)

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from agent server!")
	})
}
