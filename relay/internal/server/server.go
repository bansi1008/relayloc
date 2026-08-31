package server

import (
	"log"
	"net"
	"net/http"
	"strings"

	"relaygo/relay/internal/agent"
	agenthandle "relaygo/relay/internal/agentHandle"
	"relaygo/relay/internal/auth"
	"relaygo/relay/internal/tunnel"
	"relaygo/relay/internal/websocket"
)

type Server struct {
	mux            *http.ServeMux
	wsHandler      *websocket.Handler
	addr           string
	registry       *tunnel.Registry
	authMiddleware *auth.Middleware
}

func New(addr string, authHandler *auth.Handler, agentHandler *agenthandle.Handler, agentService *agent.Service) *Server {
	mux := http.NewServeMux()
	registry := tunnel.NewRegistry()

	s := &Server{
		addr:      addr,
		mux:       mux,
		registry:  registry,
		wsHandler: websocket.NewHandler(registry, agentService),
	}

	s.registerRoutes(authHandler, agentHandler)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	agentID := extractAgentID(r)
	if agentID != "" {
		s.handleProxy(agentID, w, r)
		return
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.addr)
	return http.ListenAndServe(s.addr, s)
}

func extractAgentID(r *http.Request) string {
	host := r.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.ToLower(strings.TrimSpace(host))

	if strings.HasSuffix(host, ".localhost") {
		sub := strings.TrimSuffix(host, ".localhost")
		if sub != "" {
			return sub
		}
	}

	parts := strings.Split(host, ".")
	if len(parts) >= 3 {
		return parts[0]
	}

	return ""
}


