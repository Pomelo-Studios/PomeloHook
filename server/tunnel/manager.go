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

func (m *Manager) Register(tunnelID, userID string, ch chan []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conns[tunnelID] = ch
	m.owners[tunnelID] = userID
}

func (m *Manager) Unregister(tunnelID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.conns, tunnelID)
	delete(m.owners, tunnelID)
}

func (m *Manager) Get(tunnelID string) (chan []byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ch, ok := m.conns[tunnelID]
	return ch, ok
}

func (m *Manager) CheckAvailable(tunnelID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if owner, ok := m.owners[tunnelID]; ok {
		return fmt.Errorf("tunnel is currently active by %s", owner)
	}
	return nil
}
