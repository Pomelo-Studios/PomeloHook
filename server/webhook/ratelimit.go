package webhook

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	rateLimitPerMinute = 60
	rateLimitBurst     = 10
	limiterTTL         = 5 * time.Minute
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiterStore is a per-IP token bucket store. Safe for concurrent use.
type RateLimiterStore struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	stop     chan struct{}
}

// NewRateLimiterStore creates a store and starts a background cleanup goroutine.
func NewRateLimiterStore() *RateLimiterStore {
	s := &RateLimiterStore{
		limiters: make(map[string]*ipLimiter),
		stop:     make(chan struct{}),
	}
	go s.cleanup()
	return s
}

// Close stops the background cleanup goroutine.
func (s *RateLimiterStore) Close() {
	close(s.stop)
}

// Allow reports whether the given IP is within its rate limit.
func (s *RateLimiterStore) Allow(ip string) bool {
	return s.get(ip).Allow()
}

func (s *RateLimiterStore) get(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.limiters[ip]; ok {
		v.lastSeen = time.Now()
		return v.limiter
	}
	l := rate.NewLimiter(rate.Every(time.Minute/rateLimitPerMinute), rateLimitBurst)
	s.limiters[ip] = &ipLimiter{limiter: l, lastSeen: time.Now()}
	return l
}

func (s *RateLimiterStore) cleanup() {
	ticker := time.NewTicker(limiterTTL)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			for ip, v := range s.limiters {
				if time.Since(v.lastSeen) > limiterTTL {
					delete(s.limiters, ip)
				}
			}
			s.mu.Unlock()
		case <-s.stop:
			return
		}
	}
}
