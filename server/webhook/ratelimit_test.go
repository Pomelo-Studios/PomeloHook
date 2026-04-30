package webhook_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	wh "github.com/pomelo-studios/pomelo-hook/server/webhook"
)

func TestRateLimiterStore_AllowsUnderLimit(t *testing.T) {
	s := wh.NewRateLimiterStore()
	defer s.Close()

	require.True(t, s.Allow("1.2.3.4"))
}

func TestRateLimiterStore_DeniesOverBurst(t *testing.T) {
	s := wh.NewRateLimiterStore()
	defer s.Close()
	ip := "5.6.7.8"

	for i := 0; i < 10; i++ {
		require.True(t, s.Allow(ip), "request %d within burst should be allowed", i+1)
	}
	require.False(t, s.Allow(ip), "request beyond burst should be denied")
}

func TestRateLimiterStore_DifferentIPsAreIndependent(t *testing.T) {
	s := wh.NewRateLimiterStore()
	defer s.Close()

	for i := 0; i < 10; i++ {
		s.Allow("10.0.0.1")
	}

	require.True(t, s.Allow("10.0.0.2"))
}
