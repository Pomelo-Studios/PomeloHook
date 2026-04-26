package tunnel

import (
	"fmt"
	"sync"
)

type Manager struct {
	mu     sync.RWMutex
	conns  map[string]chan []byte
	owners map[string]string
}

func NewManager() *Manager {
	return &Manager{
		conns:  make(map[string]chan []byte),
		owners: make(map[string]string),
	}
}

// CheckAndRegister atomically verifies the tunnel is not active and registers it.
func (m *Manager) CheckAndRegister(tunnelID, userID, userName string, ch chan []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if owner, ok := m.owners[tunnelID]; ok {
		return fmt.Errorf("tunnel is currently active by %s", owner)
	}
	m.conns[tunnelID] = ch
	m.owners[tunnelID] = userName
	return nil
}

// Unregister removes the tunnel and closes its channel. Safe to call multiple times.
func (m *Manager) Unregister(tunnelID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ch, ok := m.conns[tunnelID]; ok {
		close(ch)
		delete(m.conns, tunnelID)
		delete(m.owners, tunnelID)
	}
}

func (m *Manager) Get(tunnelID string) (chan []byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ch, ok := m.conns[tunnelID]
	return ch, ok
}
