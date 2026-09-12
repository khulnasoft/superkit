package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiterAllowsRequests(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	key := "test-ip"

	for i := 0; i < 3; i++ {
		assert.True(t, rl.Allow(key), "request %d should be allowed", i+1)
	}

	assert.False(t, rl.Allow(key), "request 4 should be denied")
}

func TestRateLimiterWindowExpiry(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)
	key := "test-ip"

	assert.True(t, rl.Allow(key))
	assert.True(t, rl.Allow(key))
	assert.False(t, rl.Allow(key))

	time.Sleep(2 * time.Second)

	assert.True(t, rl.Allow(key), "should be allowed after window expires")
}

func TestRateLimiterMiddleware(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	keyFunc := func(r *http.Request) string {
		return "test-client"
	}
	middleware := rl.Middleware(keyFunc)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := middleware(next)

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name:       "X-Forwarded-For",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.1, 198.51.100.1"},
			remoteAddr: "10.0.0.1:1234",
			expected:   "203.0.113.1",
		},
		{
			name:       "X-Real-IP",
			headers:    map[string]string{"X-Real-IP": "203.0.113.2"},
			remoteAddr: "10.0.0.1:1234",
			expected:   "203.0.113.2",
		},
		{
			name:       "RemoteAddr",
			headers:    map[string]string{},
			remoteAddr: "203.0.113.3:1234",
			expected:   "203.0.113.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			for k, v := range tt.headers {
				r.Header.Set(k, v)
			}
			r.RemoteAddr = tt.remoteAddr

			ip := GetClientIP(r)
			assert.Equal(t, tt.expected, ip)
		})
	}
}
