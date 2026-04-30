package webhook_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	wh "github.com/pomelo-studios/pomelo-hook/server/webhook"
)

func TestRateLimiterStore_AllowsUnderLimit(t *testing.T) {
	store := wh.NewRateLimiterStore()

	allowed := store.Allow("1.2.3.4")
	require.True(t, allowed)
}

func TestRateLimiterStore_DeniesOverBurst(t *testing.T) {
	store := wh.NewRateLimiterStore()
	ip := "5.6.7.8"

	allowed := 0
	denied := 0
	for i := 0; i < 20; i++ {
		if store.Allow(ip) {
			allowed++
		} else {
			denied++
		}
	}
	require.Equal(t, 10, allowed, "burst capacity should be exactly 10")
	require.Equal(t, 10, denied, "requests beyond burst should be denied")
}

func TestRateLimiterStore_DifferentIPsAreIndependent(t *testing.T) {
	store := wh.NewRateLimiterStore()

	for i := 0; i < 10; i++ {
		store.Allow("10.0.0.1")
	}

	require.True(t, store.Allow("10.0.0.2"))
}
