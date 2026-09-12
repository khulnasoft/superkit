package app

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"
	"github.com/khulnasoft/superkit/bootstrap/app/handlers"
	"github.com/khulnasoft/superkit/bootstrap/app/views/errors"
	"github.com/khulnasoft/superkit/bootstrap/kit/csrf"
	"github.com/khulnasoft/superkit/bootstrap/kit/middleware"
	"github.com/khulnasoft/superkit/bootstrap/plugins/auth"
	"github.com/khulnasoft/superkit/kit"
	kitmiddleware "github.com/khulnasoft/superkit/kit/middleware"
)

var store *sessions.CookieStore

func InitializeMiddleware(router *chi.Mux) {
	csrf.Init(store)
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := r.Cookie("csrf_token"); err != nil {
				sess, _ := store.Get(r, "csrf-session")
				csrf.SetCSRTCookie(w, r, csrf.GenerateToken(sess))
			}
			next.ServeHTTP(w, r)
		})
	})

	router.Use(middleware.CSRFMiddleware)
	router.Use(middleware.NewRateLimiter(5, time.Minute).Middleware(middleware.GetClientIP))

	if !kit.IsProduction() {
		router.Use(chimiddleware.Logger)
	}

	router.Use(chimiddleware.Recoverer)
	router.Use(kitmiddleware.WithRequest)
	router.Use(kitmiddleware.WithRequestID)
}

func InitializeHealthRoute(router *chi.Mux) {
	healthRouter := chi.NewMux()
	healthRouter.Get("/health", kit.Handler(HandleHealth))
	router.Mount("/health", healthRouter)
}

func InitializeRoutes(router *chi.Mux) {
	auth.InitializeRoutes(router)
	authConfig := kit.AuthenticationConfig{
		AuthFunc:    auth.AuthenticateUser,
		RedirectURL: "/login",
	}

	router.Group(func(app chi.Router) {
		app.Use(kit.WithAuthentication(authConfig, false))

		app.Get("/", kit.Handler(handlers.HandleLandingIndex))
	})

	router.Group(func(app chi.Router) {
		app.Use(kit.WithAuthentication(authConfig, true))

		app.Get("/profile", kit.Handler(auth.HandleProfileShow))
		app.Put("/profile", kit.Handler(auth.HandleProfileUpdate))
	})
}

func HandleHealth(kit *kit.Kit) error {
	kit.Response.Header().Set("Content-Type", "application/json")
	kit.Response.WriteHeader(http.StatusOK)
	return nil
}

func NotFoundHandler(kit *kit.Kit) error {
	return kit.Render(errors.Error404())
}

func ErrorHandler(kit *kit.Kit, err error) {
	slog.Error("internal server error", "err", err.Error(), "path", kit.Request.URL.Path)
	kit.Render(errors.Error500())
}
