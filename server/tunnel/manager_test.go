package tunnel_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func TestCheckAndRegister(t *testing.T) {
	m := tunnel.NewManager()
	ch := make(chan []byte, 1)
	err := m.CheckAndRegister("tunnel-1", "user-1", "Alice", ch)
	require.NoError(t, err)
	got, ok := m.Get("tunnel-1")
	require.True(t, ok)
	require.Equal(t, ch, got)
}

func TestCheckAndRegisterConflict(t *testing.T) {
	m := tunnel.NewManager()
	ch1 := make(chan []byte, 1)
	ch2 := make(chan []byte, 1)
	require.NoError(t, m.CheckAndRegister("tunnel-1", "user-1", "Alice", ch1))
	err := m.CheckAndRegister("tunnel-1", "user-2", "Bob", ch2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Alice")
}

func TestUnregisterClosesChannel(t *testing.T) {
	m := tunnel.NewManager()
	ch := make(chan []byte, 1)
	require.NoError(t, m.CheckAndRegister("tunnel-1", "user-1", "Alice", ch))
	m.Unregister("tunnel-1")
	_, ok := m.Get("tunnel-1")
	require.False(t, ok)
	// channel should be closed
	_, open := <-ch
	require.False(t, open)
}

func TestUnregisterIdempotent(t *testing.T) {
	m := tunnel.NewManager()
	ch := make(chan []byte, 1)
	require.NoError(t, m.CheckAndRegister("tunnel-1", "user-1", "Alice", ch))
	m.Unregister("tunnel-1")
	require.NotPanics(t, func() { m.Unregister("tunnel-1") })
}
