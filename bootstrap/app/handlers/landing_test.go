package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/khulnasoft/superkit/kit"
	"github.com/stretchr/testify/assert"
)

func TestHandleLandingIndex(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	err := HandleLandingIndex(k)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}
