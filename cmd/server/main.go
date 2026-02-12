package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bsrodrigue/appshare-backend/internal/auth"
	"github.com/bsrodrigue/appshare-backend/internal/config"
	"github.com/bsrodrigue/appshare-backend/internal/db"
	"github.com/bsrodrigue/appshare-backend/internal/handler"
	"github.com/bsrodrigue/appshare-backend/internal/handler/middleware"
	"github.com/bsrodrigue/appshare-backend/internal/logger"
	"github.com/bsrodrigue/appshare-backend/internal/repository/postgres"
	"github.com/bsrodrigue/appshare-backend/internal/service"
	"github.com/bsrodrigue/appshare-backend/internal/storage"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/jackc/pgx/v5/pgxpool"
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

	// ========== Logging ==========

	// Set up structured logger
	logCfg := logger.Config{
		Level:     cfg.LogLevel,
		Format:    cfg.LogFormat,
		Output:    "stdout",
		AddSource: cfg.Environment == "production",
	}

	if err := logger.SetDefault(logCfg); err != nil {
		log.Fatalf("Failed to set up logger: %v", err)
	}

	slog.Info("Starting AppShare API",
		slog.String("environment", cfg.Environment),
		slog.String("log_level", cfg.LogLevel),
		slog.String("log_format", cfg.LogFormat),
	)

	// Create context for database connection
	ctx := context.Background()

	// ========== Infrastructure ==========

	// Database connection pool
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Unable to parse database URL", slog.String("error", err.Error()))
		os.Exit(1)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		slog.Error("Unable to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("Unable to ping database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	slog.Info("Database connected", slog.Int("max_conns", int(poolConfig.MaxConns)))

	// sqlc queries
	queries := db.New(pool)

	// Transaction manager
	txManager := db.NewTxManager(pool)

	// JWT service
	jwtConfig := auth.JWTConfig{
		SecretKey:            cfg.JWTSecretKey,
		AccessTokenDuration:  cfg.JWTAccessTokenDuration,
		RefreshTokenDuration: cfg.JWTRefreshTokenDuration,
		Issuer:               cfg.JWTIssuer,
	}
	jwtService := auth.NewJWTService(jwtConfig)
	slog.Info("JWT configured",
		slog.Duration("access_token_duration", cfg.JWTAccessTokenDuration),
		slog.Duration("refresh_token_duration", cfg.JWTRefreshTokenDuration),
	)

	// ========== Storage ==========

	var storageSvc storage.Storage
	if cfg.R2AccountID != "" {
		storageSvc, err = storage.NewR2Storage(ctx, cfg.R2AccountID, cfg.R2AccessKeyID, cfg.R2SecretAccessKey, cfg.R2BucketName, cfg.R2PublicDomain)
		if err != nil {
			slog.Error("Failed to initialize R2 storage", slog.String("error", err.Error()))
			if cfg.Environment == "production" {
				os.Exit(1)
			}
		} else {
			slog.Info("Cloudflare R2 storage initialized", slog.String("bucket", cfg.R2BucketName))
		}
	} else {
		slog.Warn("Cloudflare R2 storage not configured (R2_ACCOUNT_ID missing)")
	}

	// ========== Repositories ==========

	userRepo := postgres.NewUserRepository(queries)
	projectRepo := postgres.NewProjectRepository(queries)
	appRepo := postgres.NewApplicationRepository(queries)
	releaseRepo := postgres.NewReleaseRepository(queries)
	artifactRepo := postgres.NewArtifactRepository(queries)

	// ========== Services ==========

	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, jwtService)
	projectService := service.NewProjectService(projectRepo, userRepo, txManager)
	appService := service.NewApplicationService(appRepo, projectRepo)
	releaseService := service.NewReleaseService(releaseRepo, appRepo, projectRepo, artifactRepo, storageSvc, txManager)
	artifactService := service.NewArtifactService(artifactRepo, releaseRepo, appRepo, projectRepo, storageSvc)
	fileService := service.NewFileService(storageSvc)

	// ========== Auth Middleware ==========

	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// ========== Router ==========

	mux := http.NewServeMux()

	// Configure Huma with security scheme
	humaConfig := huma.DefaultConfig("AppShare API", "1.0.0")
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearer": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
			Description:  "JWT access token. Get one from /auth/login or /auth/register",
		},
	}

	api := humago.New(mux, humaConfig)

	// ========== Handlers ==========

	systemHandler := handler.NewSystemHandler()
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService)
	projectHandler := handler.NewProjectHandler(projectService)
	applicationHandler := handler.NewApplicationHandler(appService)
	releaseHandler := handler.NewReleaseHandler(releaseService)
	artifactHandler := handler.NewArtifactHandler(artifactService)
	fileHandler := handler.NewFileHandler(fileService)

	// Register all routes on the main API
	systemHandler.Register(api)
	authHandler.Register(api)

	// Sub-router for protected routes - This time we'll mount it correctly
	protectedMux := http.NewServeMux()
	protectedApi := humago.New(protectedMux, humaConfig)

	authHandler.RegisterProtected(protectedApi)
	userHandler.Register(protectedApi)
	projectHandler.Register(protectedApi)
	applicationHandler.Register(protectedApi)
	releaseHandler.Register(protectedApi)
	artifactHandler.Register(protectedApi)
	fileHandler.Register(protectedApi)

	// The fix: use a catch-all route for protected routes to ensure path stripping/matching works correctly
	mux.Handle("/", authMiddleware.RequireAuth(protectedMux))

	// ========== Apply Global Middleware ==========

	loggingMiddleware := middleware.NewLoggingMiddleware(middleware.DefaultLoggingConfig())
	var rootHandler http.Handler = mux
	rootHandler = loggingMiddleware.Handler(rootHandler)

	// ========== Server ==========

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      rootHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan

		slog.Info("Shutting down server", slog.String("signal", sig.String()))

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("Server shutdown error", slog.String("error", err.Error()))
		}
	}()

	slog.Info("Server starting",
		slog.String("port", cfg.Port),
		slog.String("docs", "http://localhost:"+cfg.Port+"/docs"),
	)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("Server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("Server stopped gracefully")
}
