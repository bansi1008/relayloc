package tunnel

import (
	"sync"
)

type Registry struct {
	mu      sync.RWMutex
	tunnels map[string]*Session
}

func NewRegistry() *Registry {
	return &Registry{
		tunnels: make(map[string]*Session),
	}
}

func (r *Registry) Register(id string, s *Session) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tunnels[id] = s
}

func (r *Registry) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tunnels, id)
}

func (r *Registry) Get(id string) (*Session, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.tunnels[id]
	return s, ok
}
