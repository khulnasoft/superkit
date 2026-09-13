package view

import (
	"context"
	"net/http"
	"testing"

	"github.com/khulnasoft/superkit/kit"
	"github.com/khulnasoft/superkit/kit/middleware"
)

type testAuth struct{ authenticated bool }

func (a testAuth) Check() bool { return a.authenticated }

func TestAsset(t *testing.T) {
	if got := Asset("styles.css"); got != "/public/assets/styles.css" {
		t.Fatalf("Asset() = %q", got)
	}
}

func TestContextHelpers(t *testing.T) {
	ctx := context.Background()
	if _, ok := Auth(ctx).(kit.DefaultAuth); !ok {
		t.Fatal("Auth() without context did not return DefaultAuth")
	}
	if got := Request(ctx); got == nil || got.URL != nil {
		t.Fatal("Request() without context returned an unexpected request")
	}
	if got := URL(ctx); got != nil {
		t.Fatal("URL() without context returned a URL")
	}

	req, err := http.NewRequest(http.MethodGet, "https://example.com/profile", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx = context.WithValue(ctx, middleware.RequestKey{}, req)
	ctx = context.WithValue(ctx, kit.AuthKey{}, testAuth{authenticated: true})
	if got := Request(ctx); got != req || URL(ctx).Path != "/profile" {
		t.Fatal("request context helpers returned the wrong request")
	}
	if !Auth(ctx).Check() {
		t.Fatal("Auth() returned the wrong auth value")
	}
}
