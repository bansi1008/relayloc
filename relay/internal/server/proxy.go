package server

import (
	"encoding/json"
	"io"
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	req := protocol.HTTPReq{
		ID:     requestID,
		Method: r.Method,
		Path:   "/" + r.PathValue("path"),
		Query:  r.URL.RawQuery,
		Header: r.Header,
		Body:   body,
	}

	// for key, values := range r.Header {
	// 	if len(values) > 0 {
	// 		req.Header[key] = values[0]
	// 	}
	// }

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

	for key, values := range res.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	contentType := ""
	if values, ok := res.Header["Content-Type"]; ok && len(values) > 0 {
		contentType = values[0]
	}

	if strings.Contains(contentType, "text/html") {
		html := string(res.Body)

		prefix := "/tunnel/" + id

		html = strings.ReplaceAll(html, `href="/`, `href="`+prefix+`/`)
		html = strings.ReplaceAll(html, `src="/`, `src="`+prefix+`/`)

		res.Body = []byte(html)

		w.Header().Del("Content-Length")
	}

	w.WriteHeader(res.Status)

	_, err = w.Write(res.Body)
	if err != nil {
		log.Println("failed to write response:", err)
	}
}
