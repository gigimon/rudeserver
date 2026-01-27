package ratelimit

import (
	"sync"

	"golang.org/x/time/rate"
	"rudeserver/internal/scenario"
)

type Store struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
}

func NewStore() *Store {
	return &Store{
		limiters: make(map[string]*rate.Limiter),
	}
}

func Key(sc scenario.Scenario, clientIP string) string {
	return string(sc.Protocol) + "|" + sc.Method + "|" + sc.NormalizedPath + "|" + clientIP
}

func Allow(store *Store, sc scenario.Scenario, clientIP string) bool {
	if sc.RateLimit == nil {
		return true
	}

	key := Key(sc, clientIP)
	limiter := store.getLimiter(key, sc.RateLimit.RPS, sc.RateLimit.Burst)
	return limiter.Allow()
}

func (s *Store) getLimiter(key string, rps float64, burst int) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	if limiter, ok := s.limiters[key]; ok {
		return limiter
	}

	limiter := rate.NewLimiter(rate.Limit(rps), burst)
	s.limiters[key] = limiter
	return limiter
}
