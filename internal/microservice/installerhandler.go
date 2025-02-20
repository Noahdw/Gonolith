package microservice

import (
	"io"
	"log/slog"
	"net/http"
)

type InstallerHandler struct {
	services *Microservices
}

func NewInstallerHandler(services *Microservices) *InstallerHandler {
	return &InstallerHandler{
		services: services,
	}
}

func (h *InstallerHandler) HandleInstallMicroservice(w http.ResponseWriter, r *http.Request) {
	rawzip, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Could not create space for microservice", http.StatusInternalServerError)
		return
	}

	id, err := h.services.InstallMicroservice(rawzip)
	if err != nil {
		slog.Error("could not install microservice", "err", err.Error())
		http.Error(w, "Error installing microservice", http.StatusInternalServerError)
		return
	}
	io.WriteString(w, id)
}

func (h *InstallerHandler) HandleStopMicroservice(w http.ResponseWriter, r *http.Request) {
	serviceId := r.URL.Query().Get("id")
	err := h.services.StopMicroservice(serviceId)
	if err != nil {
		http.Error(w, "Error trying to stop microservice", http.StatusBadRequest)
	}
}

func (h *InstallerHandler) HandleStartMicroservice(w http.ResponseWriter, r *http.Request) {
	serviceId := r.URL.Query().Get("id")
	err := h.services.StartMicroservice(serviceId)
	if err != nil {
		http.Error(w, "Error trying to start microservice", http.StatusBadRequest)
	}
}
