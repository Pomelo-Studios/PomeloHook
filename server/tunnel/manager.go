package tunnel

import "sync"

type Manager struct {
	mu     sync.Mutex
	conns  map[string][]chan []byte
	owners map[string][]string
}

func NewManager() *Manager {
	return &Manager{
		conns:  make(map[string][]chan []byte),
		owners: make(map[string][]string),
	}
}

// Register adds a new subscriber for tunnelID and returns its dedicated channel.
// Always succeeds — multiple subscribers on the same tunnel are allowed.
func (m *Manager) Register(tunnelID, userName string) chan []byte {
	ch := make(chan []byte, 64)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conns[tunnelID] = append(m.conns[tunnelID], ch)
	m.owners[tunnelID] = append(m.owners[tunnelID], userName)
	return ch
}

// Unregister removes and closes the specific channel from tunnelID's subscriber list.
// Safe to call on an already-removed channel (no-op).
func (m *Manager) Unregister(tunnelID string, ch chan []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	chans := m.conns[tunnelID]
	for i, c := range chans {
		if c == ch {
			close(c)
			m.conns[tunnelID] = append(chans[:i], chans[i+1:]...)
			m.owners[tunnelID] = append(m.owners[tunnelID][:i], m.owners[tunnelID][i+1:]...)
			break
		}
	}
	if len(m.conns[tunnelID]) == 0 {
		delete(m.conns, tunnelID)
		delete(m.owners, tunnelID)
	}
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
	delete(m.owners, tunnelID)
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
