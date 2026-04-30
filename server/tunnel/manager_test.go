package tunnel_test

import (
	"bytes"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/pomelo-studios/pomelo-hook/server/tunnel"
)

func TestRegisterReturnsChannel(t *testing.T) {
	m := tunnel.NewManager()
	ch := m.Register("tunnel-1", "Alice")
	require.NotNil(t, ch)
	require.Equal(t, 1, m.SubCount("tunnel-1"))
}

func TestRegisterTwiceAllowsBoth(t *testing.T) {
	m := tunnel.NewManager()
	ch1 := m.Register("tunnel-1", "Alice")
	ch2 := m.Register("tunnel-1", "Bob")
	require.NotEqual(t, ch1, ch2)
	require.Equal(t, 2, m.SubCount("tunnel-1"))
}

func TestBroadcastFansOutToAllSubscribers(t *testing.T) {
	m := tunnel.NewManager()
	ch1 := m.Register("tunnel-1", "Alice")
	ch2 := m.Register("tunnel-1", "Bob")
	ch3 := m.Register("tunnel-1", "Carol")

	m.Broadcast("tunnel-1", []byte("hello"))

	require.Equal(t, []byte("hello"), <-ch1)
	require.Equal(t, []byte("hello"), <-ch2)
	require.Equal(t, []byte("hello"), <-ch3)
}

func TestBroadcastToEmptyListNoPanic(t *testing.T) {
	m := tunnel.NewManager()
	require.NotPanics(t, func() {
		m.Broadcast("no-such-tunnel", []byte("hello"))
	})
}

func TestUnregisterRemovesSubscriberAndClosesChannel(t *testing.T) {
	m := tunnel.NewManager()
	ch1 := m.Register("tunnel-1", "Alice")
	ch2 := m.Register("tunnel-1", "Bob")

	m.Unregister("tunnel-1", ch1)

	require.Equal(t, 1, m.SubCount("tunnel-1"))
	_, open := <-ch1
	require.False(t, open, "unregistered channel must be closed")

	// Bob's channel still works
	m.Broadcast("tunnel-1", []byte("ping"))
	require.Equal(t, []byte("ping"), <-ch2)
}

func TestUnregisterLastSubscriberDeletesKey(t *testing.T) {
	m := tunnel.NewManager()
	ch := m.Register("tunnel-1", "Alice")
	m.Unregister("tunnel-1", ch)
	require.Equal(t, 0, m.SubCount("tunnel-1"))
}

func TestUnregisterAllClosesEveryChannel(t *testing.T) {
	m := tunnel.NewManager()
	ch1 := m.Register("tunnel-1", "Alice")
	ch2 := m.Register("tunnel-1", "Bob")

	m.UnregisterAll("tunnel-1")

	require.Equal(t, 0, m.SubCount("tunnel-1"))
	_, open := <-ch1
	require.False(t, open)
	_, open = <-ch2
	require.False(t, open)
}

func TestUnregisterAllNoPanicOnMissingTunnel(t *testing.T) {
	m := tunnel.NewManager()
	require.NotPanics(t, func() {
		m.UnregisterAll("no-such-tunnel")
	})
}

func TestBroadcast_LogsWhenBufferFull(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	m := tunnel.NewManager()
	_ = m.Register("t1", "Alice") // buffer capacity = 64, don't drain it

	// Fill the buffer completely
	for i := 0; i < 64; i++ {
		m.Broadcast("t1", []byte("msg"))
	}

	// 65th broadcast hits the full buffer → should log
	m.Broadcast("t1", []byte("overflow"))

	require.Contains(t, buf.String(), "subscriber buffer full")
	require.Contains(t, buf.String(), "t1")
}

func TestBroadcastConcurrentSafe(t *testing.T) {
	m := tunnel.NewManager()
	const n = 10
	channels := make([]chan []byte, n)
	for i := 0; i < n; i++ {
		channels[i] = m.Register("tunnel-1", "user")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.Broadcast("tunnel-1", []byte("msg"))
	}()
	wg.Wait()

	for _, ch := range channels {
		select {
		case got := <-ch:
			require.Equal(t, []byte("msg"), got)
		default:
			t.Fatal("expected message in channel but got none")
		}
	}
}
