package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/khulnasoft/superkit/bootstrap/app/conf"
	"github.com/khulnasoft/superkit/bootstrap/app/db"
	"github.com/khulnasoft/superkit/bootstrap/app/views/errors"
	"github.com/khulnasoft/superkit/kit"
	"github.com/stretchr/testify/assert"
)

func TestNotFoundHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	err := NotFoundHandler(k)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestErrorHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	k := &kit.Kit{
		Response: w,
		Request:  req,
	}

	ErrorHandler(k, assert.AnError)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestError404Component(t *testing.T) {
	component := errors.Error404()
	assert.NotNil(t, component)
}

func TestError500Component(t *testing.T) {
	component := errors.Error500()
	assert.NotNil(t, component)
}

func TestPreflightRequiresValidConfigAndDatabase(t *testing.T) {
	assert.Error(t, Preflight(nil))
	assert.Error(t, Preflight(&conf.Config{Listen: "localhost"}))
}

func TestPreflightPassesWhenDatabaseIsReady(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite3")
	t.Setenv("DB_NAME", ":memory:")

	err := db.Initialize()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	assert.NoError(t, Preflight(&conf.Config{Listen: ":3000"}))
}
