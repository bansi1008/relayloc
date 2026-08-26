package server

import (
	//	"fmt"
	"net/http"
	"os"
	"relaygo/relay/internal/agentHandle"
	"relaygo/relay/internal/auth"
)

func (s *Server) registerRoutes(authHandler *auth.Handler, agentHandler *agenthandle.Handler) {
	// s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintln(w, "Hello from relay server!")
	// })
	middleware := auth.NewMiddleware(os.Getenv("JWT_SECRET"))

	s.mux.HandleFunc("/ws", s.wsHandler.Handle)
	s.mux.HandleFunc("/tunnel/{id}/{path...}", s.Proxy)
	s.mux.HandleFunc("/auth/register", authHandler.Register)
	s.mux.HandleFunc("/auth/login", authHandler.Login)
	s.mux.Handle(
		"/test",
		middleware.RequireAuth(http.HandlerFunc(authHandler.Test)),
	)

	s.mux.Handle(
		"/agentCreate",
		middleware.RequireAuth(http.HandlerFunc(agentHandler.CreateAgent)),
	)
	s.mux.Handle(
		"GET /agent/{id}",
		middleware.RequireAuth(http.HandlerFunc(agentHandler.GetAgentByID)),
	)
	s.mux.Handle(
		"GET /agents",
		middleware.RequireAuth(http.HandlerFunc(agentHandler.GetAgentsByUserID)),
	)
	s.mux.Handle(
		"DELETE /agent/{id}",
		middleware.RequireAuth(http.HandlerFunc(agentHandler.DeleteAgent)),
	)
}
