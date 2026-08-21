package server

import (
	"encoding/json"
	"log"
	"net/http"
	"relaygo/relay/internal/tunnel"
	"relaygo/shared/protocol"
	"strings"
)

func (s *Server) Proxy(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

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

	req := protocol.HTTPReq{
		ID:     requestID,
		Method: r.Method,
		Path:   "/" + r.PathValue("path"),
		Header: map[string]string{},
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

	respPayload, err := session.Request(r.Context(), requestID, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	var res protocol.HTTPRes

	if err := json.Unmarshal(respPayload, &res); err != nil {
		http.Error(w, "invalid response", http.StatusBadGateway)
		return
	}

	//log.Printf("Response Content-Type: %q", res.Header["Content-Type"])
	//log.Printf("Response headers: %+v", res.Header)

	for key, value := range res.Header {
		w.Header().Set(key, value)
	}
	if strings.Contains(res.Header["Content-Type"], "text/html") {
		html := string(res.Body)

		prefix := "/tunnel/" + id

		html = strings.ReplaceAll(html, `href="/`, `href="`+prefix+`/`)
		html = strings.ReplaceAll(html, `src="/`, `src="`+prefix+`/`)

		res.Body = []byte(html)

		// Body length has changed, so don't forward an old Content-Length.
		w.Header().Del("Content-Length")
	}

	w.WriteHeader(res.Status)

	_, err = w.Write(res.Body)
	if err != nil {
		log.Println("failed to write response:", err)
	}
}
