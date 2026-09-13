package kit

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type testAuth struct{ authenticated bool }

func (a testAuth) Check() bool { return a.authenticated }

type testComponent struct{}

func (testComponent) Render(_ context.Context, w io.Writer) error {
	_, err := w.Write([]byte("rendered"))
	return err
}

func newKit(method, target string, body *strings.Reader) (*Kit, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, target, body)
	response := httptest.NewRecorder()
	return &Kit{Response: response, Request: req}, response
}

func TestKitResponseHelpers(t *testing.T) {
	kit, response := newKit(http.MethodGet, "/", strings.NewReader(""))
	if err := kit.Text(http.StatusAccepted, "hello"); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusAccepted || response.Body.String() != "hello" {
		t.Fatalf("Text() = %d, %q", response.Code, response.Body.String())
	}

	kit, response = newKit(http.MethodGet, "/", strings.NewReader(""))
	if err := kit.Bytes(http.StatusCreated, []byte("bytes")); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusCreated || response.Body.String() != "bytes" {
		t.Fatalf("Bytes() = %d, %q", response.Code, response.Body.String())
	}

	kit, response = newKit(http.MethodGet, "/", strings.NewReader(""))
	if err := kit.JSON(http.StatusOK, map[string]string{"message": "ok"}); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusOK || response.Body.String() != "{\"message\":\"ok\"}\n" {
		t.Fatalf("JSON() = %d, %q", response.Code, response.Body.String())
	}

	kit, response = newKit(http.MethodGet, "/", strings.NewReader(""))
	if err := kit.Render(testComponent{}); err != nil {
		t.Fatal(err)
	}
	if response.Body.String() != "rendered" {
		t.Fatalf("Render() = %q", response.Body.String())
	}
}

func TestKitRequestHelpersAndAuth(t *testing.T) {
	kit, _ := newKit(http.MethodPost, "/", strings.NewReader("name=Jane"))
	kit.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if got := kit.FormValue("name"); got != "Jane" {
		t.Fatalf("FormValue() = %q; want Jane", got)
	}
	if kit.Auth().Check() {
		t.Fatal("Auth() without context should be unauthenticated")
	}
	kit.Request = kit.Request.WithContext(context.WithValue(kit.Request.Context(), AuthKey{}, testAuth{authenticated: true}))
	if !kit.Auth().Check() {
		t.Fatal("Auth() did not return the context auth")
	}
}

func TestRedirect(t *testing.T) {
	kit, response := newKit(http.MethodGet, "/from", strings.NewReader(""))
	if err := kit.Redirect(http.StatusMovedPermanently, "/to"); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusMovedPermanently || response.Header().Get("Location") != "/to" {
		t.Fatalf("Redirect() = %d, %q", response.Code, response.Header().Get("Location"))
	}

	kit, response = newKit(http.MethodGet, "/from", strings.NewReader(""))
	kit.Request.Header.Set("HX-Request", "true")
	if err := kit.Redirect(http.StatusMovedPermanently, "/to"); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusSeeOther || response.Header().Get("HX-Redirect") != "/to" {
		t.Fatalf("HTMX Redirect() = %d, %q", response.Code, response.Header().Get("HX-Redirect"))
	}
}

func TestHandlerAndErrorHandler(t *testing.T) {
	handler := Handler(func(kit *Kit) error { return kit.Text(http.StatusOK, "ok") })
	response := httptest.NewRecorder()
	handler(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusOK || response.Body.String() != "ok" {
		t.Fatalf("Handler() success = %d, %q", response.Code, response.Body.String())
	}

	oldHandler := errorHandler
	t.Cleanup(func() { errorHandler = oldHandler })
	response = httptest.NewRecorder()
	Handler(func(*Kit) error { return errors.New("broken") })(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusInternalServerError || response.Body.String() != "broken" {
		t.Fatalf("Handler() error = %d, %q", response.Code, response.Body.String())
	}

	called := false
	UseErrorHandler(func(_ *Kit, err error) {
		called = err.Error() == "custom"
	})
	Handler(func(*Kit) error { return errors.New("custom") })(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if !called {
		t.Fatal("UseErrorHandler() handler was not called")
	}
}

func TestWithAuthentication(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Context().Value(AuthKey{}).(Auth); !ok {
			t.Error("auth was not added to request context")
		}
		w.WriteHeader(http.StatusOK)
	})
	config := AuthenticationConfig{
		RedirectURL: "/login",
		AuthFunc:    func(*Kit) (Auth, error) { return testAuth{}, nil },
	}

	response := httptest.NewRecorder()
	WithAuthentication(config, true)(next).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/private", nil))
	if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "/login" {
		t.Fatalf("strict redirect = %d, %q", response.Code, response.Header().Get("Location"))
	}

	response = httptest.NewRecorder()
	WithAuthentication(config, true)(next).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/login", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("redirect target = %d; want 200", response.Code)
	}

	config.AuthFunc = func(*Kit) (Auth, error) { return testAuth{authenticated: true}, nil }
	response = httptest.NewRecorder()
	WithAuthentication(config, true)(next).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/private", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("authenticated request = %d; want 200", response.Code)
	}

	config.AuthFunc = func(*Kit) (Auth, error) { return nil, errors.New("auth failed") }
	response = httptest.NewRecorder()
	WithAuthentication(config, false)(next).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/private", nil))
	if response.Code != http.StatusInternalServerError || response.Body.String() != "auth failed" {
		t.Fatalf("auth error = %d, %q", response.Code, response.Body.String())
	}
}

func TestEnvironmentHelpers(t *testing.T) {
	t.Setenv("SUPERKIT_ENV", "development")
	if !IsDevelopment() || IsProduction() || Env() != "development" {
		t.Fatal("development environment helpers returned an unexpected result")
	}
	t.Setenv("SUPERKIT_ENV", "production")
	if IsDevelopment() || !IsProduction() || Env() != "production" {
		t.Fatal("production environment helpers returned an unexpected result")
	}
	os.Unsetenv("SUPERKIT_ENV")
	if got := Getenv("MISSING_SUPERKIT_TEST", "fallback"); got != "fallback" {
		t.Fatalf("Getenv() missing = %q", got)
	}
	t.Setenv("SUPERKIT_TEST", "configured")
	if got := Getenv("SUPERKIT_TEST", "fallback"); got != "configured" {
		t.Fatalf("Getenv() configured = %q", got)
	}
}
