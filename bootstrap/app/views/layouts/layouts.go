package layouts

import (
	"github.com/gorilla/sessions"
	"github.com/khulnasoft/superkit/bootstrap/kit/csrf"
)

var store *sessions.CookieStore

func InitCSRF(store *sessions.CookieStore) {
	csrf.Init(store)
}
