package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"viddown/config"
	"viddown/handlers"
	"viddown/middleware"
	"viddown/services"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg := config.Load()
	logger.Info("Configuration loaded",
		"port", cfg.Port,
		"authRequired", cfg.AuthRequired,
		"maxConcurrent", cfg.MaxConcurrent,
		"rateLimitRPM", cfg.RateLimitRPM,
	)

	// Initialize services
	validator := services.NewValidator()
	ytdlp := services.NewYtDlpService(cfg.YtDlpPath, validator)
	semaphore := services.NewSemaphore(cfg.MaxConcurrent)
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPM)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(cfg.YtDlpPath)
	configHandler := handlers.NewConfigHandler(cfg)
	analyzeHandler := handlers.NewAnalyzeHandler(ytdlp, logger)
	downloadHandler := handlers.NewDownloadHandler(ytdlp, semaphore, logger)
	thumbnailHandler := handlers.NewThumbnailHandler(logger)

	// Initialize router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(10 * time.Minute))

	// CORS
	r.Use(cors.Handler(middleware.CORS()))

	// Rate limiting
	r.Use(rateLimiter.Middleware)

	// Auth middleware (currently a no-op when AUTH_REQUIRED=false)
	authProvider := &middleware.NoAuthProvider{}
	r.Use(middleware.AuthMiddleware(cfg.AuthRequired, authProvider))

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", healthHandler.ServeHTTP)
		r.Get("/config", configHandler.ServeHTTP)
		r.Post("/analyze", analyzeHandler.ServeHTTP)
		r.Get("/download", downloadHandler.ServeHTTP)
		r.Get("/thumbnail", thumbnailHandler.ServeHTTP)
	})

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 10 * time.Minute, // Long timeout for downloads
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting server", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("Server stopped gracefully")
}


