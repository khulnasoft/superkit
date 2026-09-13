package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithRequest(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := r.Context().Value(RequestKey{})
		assert.NotNil(t, val)
		req, ok := val.(*http.Request)
		assert.True(t, ok, "expected *http.Request in context")
		assert.Equal(t, "/", req.URL.Path)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	WithRequest(next).ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWithRequestContextIsolation(t *testing.T) {
	key := RequestKey{}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.Context().Value(key).(*http.Request)
		assert.True(t, ok, "expected request in context")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	WithRequest(next).ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWithRequestID(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := r.Context().Value(RequestKey{}).(string)
		assert.True(t, ok, "expected request ID in context")
		assert.Len(t, id, 32, "expected a 16-byte hexadecimal request ID")
		assert.Equal(t, id, w.Header().Get("X-Request-ID"))
		w.WriteHeader(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	WithRequestID(next).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Equal(t, http.StatusNoContent, w.Code)
}
