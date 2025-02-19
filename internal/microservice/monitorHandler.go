package microservice

import (
	"fmt"
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

	var status string
	for _, service := range h.services.entries {
		config := service.GetConfig()
		status = fmt.Sprintf("%s %s, v%s", status, config.Name, config.Version)
	}
	slog.Info("Status", "running", status)
	io.WriteString(w, status)
}
