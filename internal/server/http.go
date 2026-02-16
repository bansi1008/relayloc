package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"relaygo/internal/tunnel"

	//"nhooyr.io/websocket"
)

type HTTPServer struct {
	registry *tunnel.Registry
}

func NewHTTPServer(reg *tunnel.Registry) *HTTPServer {
	return &HTTPServer{registry: reg}
}

type WSRequest struct {
	Type    string            `json:"type"`
	ID      string            `json:"id"`
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type WSResponse struct {
	Type    string            `json:"type"`
	ID      string            `json:"id"`
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

func (s *HTTPServer) HandleProxy(w http.ResponseWriter, r *http.Request) {
	
	parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/t/"), "/", 2)
	if len(parts) < 1 {
		http.Error(w, "invalid tunnel path", http.StatusBadRequest)
		return
	}

	tunnelID := parts[0]
	path := "/"
	if len(parts) == 2 {
		path = "/" + parts[1]
	}

	session, ok := s.registry.Get(tunnelID)
	if !ok {
		http.Error(w, "tunnel not connected", http.StatusBadGateway)
		return
	}

	bodyBytes, _ := io.ReadAll(r.Body)

	headers := map[string]string{}
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	req := WSRequest{
		Type:    "http_request",
		ID:      randomID(),
		Method:  r.Method,
		Path:    path,
		Headers: headers,
		Body:    base64.StdEncoding.EncodeToString(bodyBytes),
	}

	data, _ := json.Marshal(req)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

// session, ok := s.registry.Get(tunnelID)
// if !ok {
// 	http.Error(w, "tunnel not connected", http.StatusBadGateway)
// 	return
// }

respMsg, err := session.Request(ctx, req.ID, data)
if err != nil {
	http.Error(w, err.Error(), http.StatusGatewayTimeout)
	return
}

var resp WSResponse
json.Unmarshal(respMsg, &resp)


	for k, v := range resp.Headers {
		w.Header().Set(k, v)
	}

	w.WriteHeader(resp.Status)
	respBody, _ := base64.StdEncoding.DecodeString(resp.Body)
	w.Write(respBody)
}

func randomID() string {
	return time.Now().Format("150405.000000000")
}
