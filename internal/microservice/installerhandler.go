package microservice

import (
	"io"
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

	err = h.services.InstallMicroservice(rawzip)
	if err != nil {
		http.Error(w, "Error installing microservice", http.StatusInternalServerError)
	}
}
