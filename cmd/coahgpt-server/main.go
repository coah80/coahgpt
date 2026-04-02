package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/coah80/coahgpt/internal/api"
	"github.com/coah80/coahgpt/internal/auth"
	"github.com/coah80/coahgpt/internal/chat"
	"github.com/coah80/coahgpt/internal/ollama"
)

func main() {
	addr := ":8095"
	if env := os.Getenv("PORT"); env != "" {
		addr = ":" + env
	}

	if err := os.MkdirAll("./data", 0755); err != nil {
		fmt.Printf("failed to create data directory: %s\n", err)
		os.Exit(1)
	}

	authDB, err := auth.NewDB("./data/coahgpt.db")
	if err != nil {
		fmt.Printf("failed to open auth database: %s\n", err)
		os.Exit(1)
	}
	defer authDB.Close()

	var mailer auth.Mailer
	if smtpHost := os.Getenv("SMTP_HOST"); smtpHost != "" {
		mailer = auth.NewSMTPMailer(
			smtpHost,
			envOrDefault("SMTP_PORT", "1025"),
			os.Getenv("SMTP_USER"),
			os.Getenv("SMTP_PASS"),
			envOrDefault("SMTP_FROM", "noreply@coahgpt.com"),
		)
		fmt.Printf("email: using SMTP (%s:%s)\n", smtpHost, envOrDefault("SMTP_PORT", "1025"))
	} else if apiKey := os.Getenv("RESEND_API_KEY"); apiKey != "" {
		mailer = auth.NewResendMailer(apiKey)
		fmt.Println("email: using Resend API")
	} else {
		mailer = auth.NewNoopMailer()
		fmt.Println("email: using dev mailer (codes printed to stdout)")
	}

	authService := auth.NewService(authDB, mailer)
	authHandler := api.NewAuthHandler(authService)

	store := chat.NewStore()
	client := ollama.NewClient(ollama.DefaultBaseURL)
	handler := api.NewHandler(store, client, authDB, authService)

	convHandler := api.NewConversationHandler(authDB, authService)

	mux := http.NewServeMux()

	authHandler.RegisterRoutes(mux)
	convHandler.RegisterRoutes(mux)
	mux.HandleFunc("/api/chat", handler.HandleChat)
	mux.HandleFunc("/api/sessions", handler.HandleSessions)
	mux.HandleFunc("/api/health", handler.HandleHealth)

	mux.Handle("/releases/", http.StripPrefix("/releases/", http.FileServer(http.Dir("./releases"))))
	mux.Handle("/releases-v2/", http.StripPrefix("/releases-v2/", http.FileServer(http.Dir("./releases-v2"))))

	mux.HandleFunc("/install.sh", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		http.ServeFile(w, r, "./install.sh")
	})

	webDir := "./web/build"
	if info, err := os.Stat(webDir); err == nil && info.IsDir() {
		fs := http.FileServer(http.Dir(webDir))
		mux.Handle("/", spaFallback(fs, webDir))
		fmt.Printf("serving web UI from %s\n", webDir)
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"message":"coahGPT API is running. web UI not built yet."}`))
		})
		fmt.Println("web UI not found, serving API only")
	}

	// rate limit only /api/ routes, not static files
	rateLimiter := api.RateLimitMiddleware()
	finalHandler := api.LoggingMiddleware(api.CORSMiddleware(apiOnlyRateLimit(rateLimiter, mux)))

	server := &http.Server{
		Addr:         addr,
		Handler:      finalHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		fmt.Printf("coahGPT server starting on %s\n", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("server error: %s\n", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	fmt.Printf("\nreceived %s, shutting down...\n", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("shutdown error: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("server stopped")
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// spaFallback tries the file server first; if the file doesn't exist and the
// path has no extension (i.e. it's a client-side route), serve index.html.
func spaFallback(fs http.Handler, webDir string) http.Handler {
	indexPath := filepath.Join(webDir, "index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check if the file exists on disk
		fpath := filepath.Join(webDir, filepath.Clean(r.URL.Path))
		if _, err := os.Stat(fpath); err == nil {
			fs.ServeHTTP(w, r)
			return
		}
		// if it has a file extension, it's a real missing asset — 404
		if filepath.Ext(r.URL.Path) != "" {
			http.NotFound(w, r)
			return
		}
		// SPA route — serve index.html
		http.ServeFile(w, r, indexPath)
	})
}

func apiOnlyRateLimit(rl func(http.Handler) http.Handler, next http.Handler) http.Handler {
	limited := rl(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			limited.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
