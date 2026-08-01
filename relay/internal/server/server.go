package server

import (
	"log"
	"net/http"
	"relaygo/internal/websocket"
)

type Server struct {
	mux       *http.ServeMux
	wsHandler *websocket.Handler
	addr      string
}

func New(addr string) *Server {
	mux := http.NewServeMux()

	s := &Server{
		addr:      addr,
		mux:       mux,
		wsHandler: websocket.NewHandler(),
	}

	s.registerRoutes()

	return s
}

func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.addr)
	return http.ListenAndServe(s.addr, s.mux)
}
