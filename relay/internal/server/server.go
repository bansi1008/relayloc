package server

import (
	"log"
	"net/http"
	"relaygo/relay/internal/tunnel"
	"relaygo/relay/internal/websocket"
)

type Server struct {
	mux       *http.ServeMux
	wsHandler *websocket.Handler
	addr      string
	registry  *tunnel.Registry
}

func New(addr string) *Server {
	mux := http.NewServeMux()
	registry := tunnel.NewRegistry()

	s := &Server{
		addr:      addr,
		mux:       mux,
		registry:  registry,
		wsHandler: websocket.NewHandler(registry),
	}

	s.registerRoutes()

	return s
}

func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.addr)
	return http.ListenAndServe(s.addr, s.mux)
}
