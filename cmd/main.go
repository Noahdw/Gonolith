package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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

	memberPortStr := os.Getenv("MEMBERLIST_PORT")
	if memberPortStr == "" {
		memberPortStr = "7946"
	}
	memberPort, err := strconv.Atoi(memberPortStr)
	if err != nil {
		panic(err)
	}

	// Configure memberlist
	config := memberlist.DefaultLocalConfig()
	config.Name = nodeName
	config.BindPort = memberPort
	config.AdvertisePort = memberPort

	list, err := memberlist.Create(config)
	if err != nil {
		panic("Failed to create memberlist: " + err.Error())
	}

	// Join cluster if CLUSTER_MEMBERS is set
	if members := os.Getenv("CLUSTER_MEMBERS"); members != "" {
		memberList := strings.Split(members, ",")
		err := joinClusterWithRetry(list, memberList, 5, time.Second*3)
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

func joinClusterWithRetry(list *memberlist.Memberlist, members []string, retries int, delay time.Duration) error {
	var lastErr error
	for i := 0; i < retries; i++ {
		_, err := list.Join(members)
		if err == nil {
			return nil
		}
		lastErr = err
		slog.Warn("Failed to join cluster, retrying...", "attempt", i+1, "error", err)
		time.Sleep(delay)
	}
	return fmt.Errorf("failed to join cluster after %d attempts: %v", retries, lastErr)
}
