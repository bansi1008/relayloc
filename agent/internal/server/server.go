package server

import (
	"log"
	"net/http"
)

type Server struct {
	mux  *http.ServeMux
	addr string
}

func New(addr string) *Server {
	mux := http.NewServeMux()

	s := &Server{
		addr: addr,
		mux:  mux,
	}

	s.registerRoutes()

	return s
}

func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.addr)
	return http.ListenAndServe(s.addr, s.mux)
}
