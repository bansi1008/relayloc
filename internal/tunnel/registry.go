package tunnel

import (
	"sync"

	"nhooyr.io/websocket"
)

type Registry struct {
	mu      sync.RWMutex
	tunnels map[string]*websocket.Conn
}

func NewRegistry() *Registry {
	return &Registry{
		tunnels: make(map[string]*websocket.Conn),
	}
}

func (r *Registry) Register(id string, conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tunnels[id] = conn
}

func (r *Registry) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tunnels, id)
}

func (r *Registry) Get(id string) (*websocket.Conn, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	conn, ok := r.tunnels[id]
	return conn, ok
}
