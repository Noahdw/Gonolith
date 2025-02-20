package main

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/hashicorp/memberlist"
	"github.com/noahdw/Gonolith/internal/microservice"
)

func main() {

	// Get configuration from environment
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		nodeName = "gonolith1"
	}

	// Configure memberlist
	config := memberlist.DefaultLocalConfig()
	config.Name = nodeName

	list, err := memberlist.Create(config)
	if err != nil {
		panic("Failed to create memberlist: " + err.Error())
	}

	// Join cluster if CLUSTER_MEMBERS is set
	if members := os.Getenv("CLUSTER_MEMBERS"); members != "" {
		memberList := strings.Split(members, ",")
		_, err = list.Join(memberList)
		if err != nil {
			panic("Failed to join cluster: " + err.Error())
		}
	}

	services := microservice.NewMicroservices()
	handler := microservice.NewInstallerHandler(services)
	monitorHandler := microservice.NewMonitorHandler(services)
	r := chi.NewMux()
	r.Post("/install-service", handler.HandleInstallMicroservice)
	r.Post("/stop-service", handler.HandleStopMicroservice)
	r.Post("/start-service", handler.HandleStartMicroservice)
	r.Get("/get-status", monitorHandler.HandleGetStatus)

	// Create and start health checker
	checker := microservice.NewHealthChecker(services)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go checker.Start(ctx)

	http.ListenAndServe("0.0.0.0:"+httpPort, r)
}
