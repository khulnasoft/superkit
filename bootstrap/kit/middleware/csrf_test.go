package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/khulnasoft/superkit/bootstrap/kit/csrf"
	"github.com/stretchr/testify/assert"
)

func TestCSRFMiddlewareGETAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFMiddlewarePOSTWithoutToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRFMiddlewarePOSTWithValidToken(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test-secret"))
	csrf.Init(store)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	sess, _ := store.Get(req, "test")
	token := csrf.GenerateToken(sess)
	csrf.SetCSRTCookie(w, req, token)

	req.Header.Set("X-CSRF-Token", token)
	for _, cookie := range w.Result().Cookies() {
		req.AddCookie(cookie)
	}

	w2 := httptest.NewRecorder()
	CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w2, req)

	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestCSRFMiddlewarePOSTWithInvalidToken(t *testing.T) {
	store := sessions.NewCookieStore([]byte("test-secret"))
	csrf.Init(store)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	csrf.SetCSRTCookie(w, req, "valid-token")
	req.Header.Set("X-CSRF-Token", "invalid-token")
	for _, cookie := range w.Result().Cookies() {
		req.AddCookie(cookie)
	}

	w2 := httptest.NewRecorder()
	CSRFMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w2, req)

	assert.Equal(t, http.StatusForbidden, w2.Code)
}
