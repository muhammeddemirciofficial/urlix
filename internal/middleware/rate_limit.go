package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type visitor struct {
	count       int
	windowStart time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		now := time.Now()

		rl.mu.Lock()

		for ip, v := range rl.visitors {
			if now.Sub(v.windowStart) >= rl.window {
				delete(rl.visitors, ip)
			}
		}

		v, exists := rl.visitors[ip]

		if !exists || now.Sub(v.windowStart) >= rl.window {
			rl.visitors[ip] = &visitor{
				count:       0,
				windowStart: now,
			}

			rl.mu.Unlock()

			next.ServeHTTP(w, r)
			return
		}

		if v.count >= rl.limit {
			rl.mu.Unlock()

			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		v.count++

		rl.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
