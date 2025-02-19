package main

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/noahdw/Gonolith/internal/microservice"
)

func main() {
	r := chi.NewMux()

	services := microservice.Microservices{}
	handler := microservice.NewInstallerHandler(&services)

	r.Post("/install-service", handler.HandleInstallMicroservice)

	// Create and start health checker
	checker := microservice.NewHealthChecker(&services)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go checker.Start(ctx)

	http.ListenAndServe(":8080", r)
}
