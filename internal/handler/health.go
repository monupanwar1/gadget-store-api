package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type healthHandler struct {
	log *slog.Logger
}

func NewHealthHandler(log *slog.Logger) *healthHandler {
	return &healthHandler{
		log: log,
	}

}

type HealthResponse struct {
	Status  int        `json:"status"`
	Message string     `json:"message"`
	Data    HealthData `json:"data"`
}

type HealthData struct {
	Status string `json:"status"`
}

func (h *healthHandler) Health(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	response := HealthResponse{
		Status:  http.StatusOK,
		Message: "server is healthy",
		Data: HealthData{
			Status: "ok",
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.log.Error("failed to encode health response", "error", err)
	}
}
