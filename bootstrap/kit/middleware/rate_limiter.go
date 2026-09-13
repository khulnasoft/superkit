package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const shardCount = 256

type clientCounter struct {
	count    int
	windowStart time.Time
}

type shard struct {
	mu       sync.Mutex
	counters map[string]clientCounter
}

type RateLimiter struct {
	shards []shard
	limit  int
	window time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		shards: make([]shard, shardCount),
		limit:  limit,
		window: window,
	}
	for i := range rl.shards {
		rl.shards[i].counters = make(map[string]clientCounter)
	}
	return rl
}

func (rl *RateLimiter) shardForKey(key string) *shard {
	h := 0
	for i := range key {
		h = 31*h + int(key[i])
	}
	return &rl.shards[h&(shardCount-1)]
}

func (rl *RateLimiter) Allow(key string) bool {
	s := rl.shardForKey(key)
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	c, exists := s.counters[key]
	if !exists || now.Sub(c.windowStart) >= rl.window {
		s.counters[key] = clientCounter{
			count:       1,
			windowStart: now,
		}
		return true
	}

	if c.count >= rl.limit {
		return false
	}

	c.count++
	s.counters[key] = c
	return true
}

func (rl *RateLimiter) Middleware(keyFunc func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
			if !rl.Allow(key) {
				w.Header().Set("Retry-After", "60")
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		token := r.Header.Get("X-CSRF-Token")
		cookie, err := r.Cookie("csrf_token")
		if err != nil || token == "" || token != cookie.Value {
			w.Header().Set("HX-Retarget", "#error-message")
			w.Header().Set("HX-Reswap", "innerHTML")
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, "Invalid CSRF token")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func GetClientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.Split(fwd, ",")[0]
	}
	if fwd := r.Header.Get("X-Real-IP"); fwd != "" {
		return fwd
	}
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		return ip[:idx]
	}
	return ip
}

func (rl *RateLimiter) CleanupExpiredSessions(db interface{}) {
	_ = db
}
