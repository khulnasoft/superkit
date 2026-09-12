package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/khulnasoft/superkit/bootstrap/app"
	"github.com/khulnasoft/superkit/bootstrap/app/conf"
	"github.com/khulnasoft/superkit/bootstrap/app/db"
	"github.com/khulnasoft/superkit/bootstrap/public"
	"github.com/khulnasoft/superkit/kit"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("warning: could not load .env file: %v\n", err)
	}

	cfg, err := conf.Load()
	if err != nil {
		fmt.Printf("fatal: configuration error: %v\n", err)
		os.Exit(1)
	}

	kit.Setup()

	if err := db.Initialize(); err != nil {
		fmt.Printf("fatal: database initialization failed: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	router := chi.NewMux()

	app.InitializeHealthRoute(router)

	app.InitializeMiddleware(router)

	if kit.IsDevelopment() {
		router.Handle("/public/*", disableCache(staticDev()))
	} else if kit.IsProduction() {
		router.Handle("/public/*", staticProd())
	}

	kit.UseErrorHandler(app.ErrorHandler)
	router.HandleFunc("/*", kit.Handler(app.NotFoundHandler))

	app.InitializeRoutes(router)
	app.RegisterEvents()

	url := "http://localhost:7331"
	if kit.IsProduction() {
		url = fmt.Sprintf("http://localhost%s", cfg.Listen)
	}

	server := &http.Server{
		Addr:              cfg.Listen,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	go func() {
		fmt.Printf("application running in %s at %s\n", kit.Env(), url)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("fatal: server error: %v\n", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("server shutdown error: %v\n", err)
	}
}

func staticDev() http.Handler {
	return http.StripPrefix("/public/", http.FileServerFS(os.DirFS("public")))
}

func staticProd() http.Handler {
	h := http.StripPrefix("/public/", http.FileServerFS(public.AssetsFS))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Header().Set("Vary", "Accept-Encoding")
		h.ServeHTTP(w, r)
	})
}

func disableCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
