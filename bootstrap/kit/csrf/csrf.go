package csrf

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

func Init(s *sessions.CookieStore) {
	store = s
}

const csrfTokenLength = 32

func GenerateToken(session *sessions.Session) string {
	if session != nil {
		if session.Values["csrf_token"] == nil {
			session.Values["csrf_token"] = randomString(csrfTokenLength)
		}
		return session.Values["csrf_token"].(string)
	}
	return randomString(csrfTokenLength)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}

func SetCSRTCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   r.TLS != nil,
	})
}
