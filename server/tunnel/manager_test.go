package tunnel_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func TestRegisterAndGet(t *testing.T) {
	m := tunnel.NewManager()
	ch := make(chan []byte, 1)
	m.Register("tunnel-1", "user-1", "Alice", ch)
	got, ok := m.Get("tunnel-1")
	require.True(t, ok)
	require.Equal(t, ch, got)
}

func TestUnregister(t *testing.T) {
	m := tunnel.NewManager()
	ch := make(chan []byte, 1)
	m.Register("tunnel-1", "user-1", "Alice", ch)
	m.Unregister("tunnel-1")
	_, ok := m.Get("tunnel-1")
	require.False(t, ok)
}

func TestOrgTunnelConflictReturnsActiveUser(t *testing.T) {
	m := tunnel.NewManager()
	ch := make(chan []byte, 1)
	m.Register("tunnel-1", "user-1", "Alice", ch)
	err := m.CheckAvailable("tunnel-1")
	require.Error(t, err)
	require.Contains(t, err.Error(), "Alice")
}
