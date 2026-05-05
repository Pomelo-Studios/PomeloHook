package tunnel

import (
	"log"
	"sync"
)

type Manager struct {
	mu      sync.Mutex
	conns   map[string][]chan []byte
	streams map[string][]chan []byte
}

func NewManager() *Manager {
	return &Manager{
		conns:   make(map[string][]chan []byte),
		streams: make(map[string][]chan []byte),
	}
}

// Multiple subscribers on the same tunnel are allowed.
func (m *Manager) Register(tunnelID, _ string) chan []byte {
	ch := make(chan []byte, 64)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conns[tunnelID] = append(m.conns[tunnelID], ch)
	return ch
}

// Returns true if this was the last subscriber (determined atomically under the lock).
// Safe to call on an already-removed channel (no-op, returns false).
func (m *Manager) Unregister(tunnelID string, ch chan []byte) (wasLast bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	chans := m.conns[tunnelID]
	for i, c := range chans {
		if c == ch {
			close(c)
			m.conns[tunnelID] = append(chans[:i], chans[i+1:]...)
			break
		}
	}
	if len(m.conns[tunnelID]) == 0 {
		delete(m.conns, tunnelID)
		return true
	}
	return false
}

// Used by admin disconnect/delete operations.
func (m *Manager) UnregisterAll(tunnelID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ch := range m.conns[tunnelID] {
		close(ch)
	}
	delete(m.conns, tunnelID)
	for _, ch := range m.streams[tunnelID] {
		close(ch)
	}
	delete(m.streams, tunnelID)
}

// Non-blocking per subscriber: drops the message if a channel's buffer is full.
func (m *Manager) Broadcast(tunnelID string, payload []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ch := range m.conns[tunnelID] {
		select {
		case ch <- payload:
		default:
			log.Printf("tunnel %s: subscriber buffer full, event dropped", tunnelID)
		}
	}
}

func (m *Manager) SubCount(tunnelID string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.conns[tunnelID])
}

func (m *Manager) StreamCount(tunnelID string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.streams[tunnelID])
}

func (m *Manager) RegisterStream(tunnelID string) chan []byte {
	ch := make(chan []byte, 64)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.streams[tunnelID] = append(m.streams[tunnelID], ch)
	return ch
}

// Safe to call on an already-removed channel (no-op).
func (m *Manager) UnregisterStream(tunnelID string, ch chan []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	subs := m.streams[tunnelID]
	for i, c := range subs {
		if c == ch {
			close(c)
			m.streams[tunnelID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	if len(m.streams[tunnelID]) == 0 {
		delete(m.streams, tunnelID)
	}
}

// Non-blocking per subscriber: drops the message if a channel's buffer is full.
func (m *Manager) BroadcastEvent(tunnelID string, eventJSON []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ch := range m.streams[tunnelID] {
		select {
		case ch <- eventJSON:
		default:
			log.Printf("tunnel %s: stream subscriber buffer full, event dropped", tunnelID)
		}
	}
}
