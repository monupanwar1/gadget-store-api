package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type HealthHandler struct {
	log *slog.Logger
}

func NewHealthHandler(log *slog.Logger) *HealthHandler {
	return &HealthHandler{
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

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {

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
