package server

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"relaygo/relay/internal/tunnel"
	"relaygo/shared/protocol"
)

func (s *Server) Proxy(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s.handleProxy(id, w, r)
}

func (s *Server) handleProxy(id string, w http.ResponseWriter, r *http.Request) {
	session, ok := s.registry.Get(id)
	if !ok {
		http.Error(w, "agent not connected", http.StatusBadGateway)
		return
	}

	requestID, err := tunnel.NewID()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	path := r.URL.Path
	if path == "" {
		path = "/"
	}

	headers := make(map[string][]string)
	for k, v := range r.Header {
		lowerK := strings.ToLower(k)
		if lowerK == "connection" || lowerK == "keep-alive" || lowerK == "transfer-encoding" || lowerK == "upgrade" {
			continue
		}
		headers[k] = v
	}

	req := protocol.HTTPReq{
		ID:     requestID,
		Method: r.Method,
		Path:   path,
		Query:  r.URL.RawQuery,
		Header: headers,
		Body:   body,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := protocol.Encode(protocol.Frame{
		Type:    protocol.HTTPReqFrame,
		Payload: payload,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	respPayload, err := session.Request(ctx, requestID, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	var res protocol.HTTPRes
	if err := json.Unmarshal(respPayload, &res); err != nil {
		http.Error(w, "invalid response", http.StatusBadGateway)
		return
	}

	for key, values := range res.Header {
		lowerKey := strings.ToLower(key)
		if lowerKey == "connection" || lowerKey == "keep-alive" || lowerKey == "transfer-encoding" {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(res.Status)

	_, err = w.Write(res.Body)
	if err != nil {
		log.Println("failed to write response:", err)
	}
}

