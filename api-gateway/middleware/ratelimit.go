package middleware

import (
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen int64
}

type RateLimiter struct {
	mu       sync.RWMutex
	ipLimit  map[string]*ipLimiter
	rate     rate.Limit
	burst    int
}

func NewRateLimiter(rps int, burst int) *RateLimiter {
	return &RateLimiter{
		ipLimit: make(map[string]*ipLimiter),
		rate:    rate.Limit(rps),
		burst:   burst,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		rl.mu.RLock()
		entry, exists := rl.ipLimit[ip]
		rl.mu.RUnlock()

		if !exists {
			entry = &ipLimiter{limiter: rate.NewLimiter(rl.rate, rl.burst)}
			rl.mu.Lock()
			rl.ipLimit[ip] = entry
			rl.mu.Unlock()
		}

		if !entry.limiter.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
