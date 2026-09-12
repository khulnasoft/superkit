package app

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"
	"github.com/khulnasoft/superkit/bootstrap/app/conf"
	"github.com/khulnasoft/superkit/bootstrap/app/db"
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
	healthRouter.Get("/health/ready", kit.Handler(HandleReadiness))
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
	return writeStatus(kit, http.StatusOK, map[string]string{
		"status": "ok",
		"env":    kit.Getenv("SUPERKIT_ENV", "development"),
	})
}

func HandleReadiness(kit *kit.Kit) error {
	env := kit.Getenv("SUPERKIT_ENV", "development")
	if !db.IsReady() {
		return writeStatus(kit, http.StatusServiceUnavailable, map[string]string{
			"status": "error",
			"env":    env,
			"reason": "database_not_ready",
		})
	}

	return writeStatus(kit, http.StatusOK, map[string]string{
		"status": "ok",
		"env":    env,
	})
}

func Preflight(cfg *conf.Config) error {
	if cfg == nil {
		return fmt.Errorf("configuration is required")
	}
	if cfg.Listen == "" || cfg.Listen == ":" {
		return fmt.Errorf("invalid configuration: HTTP_LISTEN_ADDR is required")
	}
	if cfg.Listen != ":3000" && cfg.Listen != "localhost:3000" && cfg.Listen != "127.0.0.1:3000" {
		if cfg.Listen != ":" && len(cfg.Listen) > 0 && cfg.Listen[0] != ':' {
			for i := 0; i < len(cfg.Listen); i++ {
				if cfg.Listen[i] == ':' {
					return nil
				}
			}
			return fmt.Errorf("invalid configuration: HTTP_LISTEN_ADDR must be a valid host:port")
		}
	}
	if !db.IsReady() {
		return fmt.Errorf("database is not ready")
	}
	return nil
}

func writeStatus(kit *kit.Kit, status int, payload map[string]string) error {
	kit.Response.Header().Set("Content-Type", "application/json")
	kit.Response.WriteHeader(status)
	return json.NewEncoder(kit.Response).Encode(payload)
}

func NotFoundHandler(kit *kit.Kit) error {
	return kit.Render(errors.Error404())
}

func ErrorHandler(kit *kit.Kit, err error) {
	slog.Error("internal server error", "err", err.Error(), "path", kit.Request.URL.Path)
	kit.Render(errors.Error500())
}
