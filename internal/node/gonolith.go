package node

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

type gonolith struct {
}

func NewGonolith() *gonolith {
	return &gonolith{}
}

// Start gonolith
func (g *gonolith) Serve() {

	list, err := memberlist.Create(getMemberlistConfig())
	if err != nil {
		panic("Failed to create memberlist: " + err.Error())
	}

	// Join cluster of other gonolith nodes
	if members := os.Getenv("CLUSTER_MEMBERS"); members != "" {
		memberList := strings.Split(members, ",")
		err := joinClusterWithRetry(list, memberList, 5, time.Second*3)
		if err != nil {
			panic("Failed to join cluster: " + err.Error())
		}
	}

	// Primary data used for most functioning of a node / cluster
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

	http.ListenAndServe("0.0.0.0:"+GetHttpPort(), r)
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

func getMemberlistConfig() *memberlist.Config {
	// Get configuration from environment
	nodeName := GetNodeName()
	memberPortStr := GetMemberPort()

	memberPort, err := strconv.Atoi(memberPortStr)
	if err != nil {
		panic(err)
	}

	// Configure memberlist
	config := memberlist.DefaultLocalConfig()
	config.Name = nodeName
	config.BindPort = memberPort
	config.AdvertisePort = memberPort
	return config
}
