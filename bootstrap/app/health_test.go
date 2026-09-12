package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/khulnasoft/superkit/kit"
	"github.com/stretchr/testify/assert"
)

func TestHandleHealth(t *testing.T) {
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
}
