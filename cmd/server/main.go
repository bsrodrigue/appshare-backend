package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/bsrodrigue/appshare-backend/internal/db"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

// ApiResponse defines the general structure for all API responses
type ApiResponse[T any] struct {
	Status  int    `json:"status" doc:"HTTP status code"`
	Message string `json:"message" doc:"Brief description of the response"`
	Data    T      `json:"data" doc:"The actual response payload"`
}

// UserListResponse is the specific response for listing users
type UserListResponse struct {
	Body ApiResponse[[]db.User]
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		user := os.Getenv("POSTGRES_USER")
		pass := os.Getenv("POSTGRES_PASSWORD")
		dbName := os.Getenv("POSTGRES_DB")
		host := os.Getenv("POSTGRES_HOST")
		if host == "" {
			host = "localhost"
		}
		port := os.Getenv("POSTGRES_PORT")
		if port == "" {
			port = "5432"
		}
		dbURL = "postgres://" + user + ":" + pass + "@" + host + ":" + port + "/" + dbName + "?sslmode=disable"
	}

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	queries := db.New(conn)

	// Create a new mux and wrap it with Huma
	mux := http.NewServeMux()
	api := humago.New(mux, huma.DefaultConfig("AppShare API", "1.0.0"))

	// Register /health endpoint
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health Check",
		Description: "Verify the service is up and running.",
		Tags:        []string{"System"},
	}, func(ctx context.Context, input *struct{}) (*ApiResponse[string], error) {
		return &ApiResponse[string]{
			Status:  http.StatusOK,
			Message: "Service is healthy",
			Data:    "ok",
		}, nil
	})

	// Register /users endpoint
	huma.Register(api, huma.Operation{
		OperationID: "list-users",
		Method:      http.MethodGet,
		Path:        "/users",
		Summary:     "List Users",
		Description: "Retrieve a list of all active (non-deleted) users.",
		Tags:        []string{"Users"},
	}, func(ctx context.Context, input *struct{}) (*UserListResponse, error) {
		users, err := queries.ListUsers(ctx)
		if err != nil {
			return nil, huma.Error500InternalServerError("Failed to fetch users", err)
		}

		return &UserListResponse{
			Body: ApiResponse[[]db.User]{
				Status:  http.StatusOK,
				Message: "Users retrieved successfully",
				Data:    users,
			},
		}, nil
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on :%s", port)
	log.Printf("Documentation available at http://localhost:%s/docs", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
