package tunnel

import "sync"

type Manager struct {
	mu    sync.Mutex
	conns map[string][]chan []byte
}

func NewManager() *Manager {
	return &Manager{
		conns: make(map[string][]chan []byte),
	}
}

// Register adds a new subscriber for tunnelID and returns its dedicated channel.
// Always succeeds — multiple subscribers on the same tunnel are allowed.
func (m *Manager) Register(tunnelID, _ string) chan []byte {
	ch := make(chan []byte, 64)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conns[tunnelID] = append(m.conns[tunnelID], ch)
	return ch
}

// Unregister removes and closes the specific channel from tunnelID's subscriber list.
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

// UnregisterAll closes and removes every subscriber for tunnelID.
// Used by admin disconnect/delete operations.
func (m *Manager) UnregisterAll(tunnelID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ch := range m.conns[tunnelID] {
		close(ch)
	}
	delete(m.conns, tunnelID)
}

// Broadcast sends payload to all subscribers of tunnelID.
// Non-blocking per subscriber: drops the message if a channel's buffer is full.
func (m *Manager) Broadcast(tunnelID string, payload []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ch := range m.conns[tunnelID] {
		select {
		case ch <- payload:
		default:
		}
	}
}

// SubCount returns the number of active subscribers for tunnelID.
func (m *Manager) SubCount(tunnelID string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.conns[tunnelID])
}
