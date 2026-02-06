package handler

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// SystemHandler handles system-related HTTP requests.
type SystemHandler struct{}

// NewSystemHandler creates a new SystemHandler.
func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

// Register registers all system routes with the API.
func (h *SystemHandler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health Check",
		Description: "Verify the service is up and running.",
		Tags:        []string{"System"},
	}, h.healthCheck)
}

// HealthOutput is the response for health check.
type HealthOutput struct {
	Body ApiResponse[string]
}

func (h *SystemHandler) healthCheck(ctx context.Context, input *struct{}) (*HealthOutput, error) {
	return &HealthOutput{
		Body: ok("Service is healthy", "ok"),
	}, nil
}
