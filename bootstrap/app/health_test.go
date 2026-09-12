package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/khulnasoft/superkit/kit"
	"github.com/stretchr/testify/assert"
)

func TestHandleHealth(t *testing.T) {
	t.Setenv("SUPERKIT_ENV", "development")
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	err := HandleHealth(k)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"status":"ok","env":"development"}`, w.Body.String())
}

func TestHandleReadiness(t *testing.T) {
	t.Setenv("SUPERKIT_ENV", "production")
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	err := HandleReadiness(k)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"status":"error","env":"production","reason":"database_not_ready"}`, w.Body.String())
}
