package microservice

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

type MonitorHandler struct {
	services *Microservices
}

func NewMonitorHandler(services *Microservices) *MonitorHandler {
	return &MonitorHandler{
		services: services,
	}
}

func (h *MonitorHandler) HandleGetStatus(w http.ResponseWriter, r *http.Request) {
	if len(h.services.entries) == 0 {
		io.WriteString(w, "No services installed")
		return
	}

	statuses := h.services.GetAllStatuses()
	slog.Info("Services status:", "services", statuses)
	json.NewEncoder(w).Encode(statuses)
}
