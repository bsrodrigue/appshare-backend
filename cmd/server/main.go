// Package main is the entry point for the AppShare API server.
// This file should ONLY handle:
//   - Loading configuration
//   - Creating infrastructure (database connections)
//   - Wiring up dependencies
//   - Starting the server
//
// Business logic belongs in internal/service.
// HTTP handling belongs in internal/handler.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bsrodrigue/appshare-backend/internal/config"
	"github.com/bsrodrigue/appshare-backend/internal/db"
	"github.com/bsrodrigue/appshare-backend/internal/handler"
	"github.com/bsrodrigue/appshare-backend/internal/repository/postgres"
	"github.com/bsrodrigue/appshare-backend/internal/service"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file (optional - allows running without it in production)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	// Load and validate configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create context for database connection
	ctx := context.Background()

	// ========== Infrastructure ==========

	// Database connection
	conn, err := pgx.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// Verify database connection
	if err := conn.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}
	log.Println("✓ Database connected")

	// sqlc queries
	queries := db.New(conn)

	// ========== Repositories ==========

	userRepo := postgres.NewUserRepository(queries)
	projectRepo := postgres.NewProjectRepository(queries)

	// ========== Services ==========

	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo)
	projectService := service.NewProjectService(projectRepo, userRepo)

	// Silence unused variable warning (will be used when we add project handlers)
	_ = projectService

	// ========== Handlers ==========

	systemHandler := handler.NewSystemHandler()
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService)

	// ========== Router ==========

	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("AppShare API", "1.0.0"))

	// Register all routes
	systemHandler.Register(api)
	userHandler.Register(api)
	authHandler.Register(api)

	// ========== Server ==========

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("✓ Server starting on :%s", cfg.Port)
	log.Printf("✓ Documentation available at http://localhost:%s/docs", cfg.Port)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}

	log.Println("Server stopped gracefully")
}
