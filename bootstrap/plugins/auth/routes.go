package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/khulnasoft/superkit/kit"
)

func RequireRole(role Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth, ok := r.Context().Value(kit.AuthKey{}).(Auth)
			if !ok || !auth.Check() || !auth.HasRole(role) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth, ok := r.Context().Value(kit.AuthKey{}).(Auth)
			if !ok || !auth.Check() || !auth.Can(permission) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func InitializeRoutes(router chi.Router) {
	authConfig := kit.AuthenticationConfig{
		AuthFunc:    AuthenticateUser,
		RedirectURL: "/login",
	}

	router.Get("/email/verify", kit.Handler(HandleEmailVerify))
	router.Post("/resend-email-verification", kit.Handler(HandleResendVerificationCode))

	router.Group(func(auth chi.Router) {
		auth.Use(kit.WithAuthentication(authConfig, false))
		auth.Get("/login", kit.Handler(HandleLoginIndex))
		auth.Post("/login", kit.Handler(HandleLoginCreate))
		auth.Delete("/logout", kit.Handler(HandleLoginDelete))

		auth.Get("/signup", kit.Handler(HandleSignupIndex))
		auth.Post("/signup", kit.Handler(HandleSignupCreate))
	})

	router.With(kit.WithAuthentication(authConfig, true), RequireRole(RoleAdmin)).Get("/admin", kit.Handler(HandleAdminIndex))
}
