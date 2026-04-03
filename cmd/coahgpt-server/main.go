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
	"github.com/coah80/coahgpt/internal/chat"
	"github.com/coah80/coahgpt/internal/ollama"
)

func main() {
	addr := ":8095"
	if env := os.Getenv("PORT"); env != "" {
		addr = ":" + env
	}

	ollamaURL := ollama.DefaultBaseURL
	if env := os.Getenv("OLLAMA_URL"); env != "" {
		ollamaURL = env
	}

	store := chat.NewStore()
	client := ollama.NewClient(ollamaURL)
	handler := api.NewHandler(store, client)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/chat", handler.HandleChat)
	mux.HandleFunc("/api/sessions", handler.HandleSessions)
	mux.HandleFunc("/api/health", handler.HandleHealth)

	// OpenAI-compatible API
	mux.HandleFunc("/v1/chat/completions", api.HandleOpenAIChatCompletions)
	mux.HandleFunc("/v1/models", api.HandleOpenAIModels)

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

func spaFallback(fs http.Handler, webDir string) http.Handler {
	indexPath := filepath.Join(webDir, "index.html")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fpath := filepath.Join(webDir, filepath.Clean(r.URL.Path))
		if _, err := os.Stat(fpath); err == nil {
			fs.ServeHTTP(w, r)
			return
		}
		if filepath.Ext(r.URL.Path) != "" {
			http.NotFound(w, r)
			return
		}
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
