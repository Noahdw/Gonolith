package microservice

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type HealthChecker struct {
	services      *Microservices
	checkInterval time.Duration
}

func NewHealthChecker(services *Microservices) *HealthChecker {
	return &HealthChecker{
		services:      services,
		checkInterval: 10 * time.Second,
	}
}

func (h *HealthChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(h.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.checkServices()
		}
	}
}

func (h *HealthChecker) checkServices() {
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	// TODO: Make concurrent
	for _, service := range h.services.entries {

		//TODO: Better way to do status, maybe enum
		if service.status != "running" {
			continue
		}

		// Create connection to service's gRPC server
		conn, err := grpc.Dial(fmt.Sprintf("localhost:%s", grpcPort), grpc.WithInsecure())
		if err != nil {
			slog.Error("Failed to connect to service", "name", service.exeFileName, "error", err)
			continue
		}
		defer conn.Close()

		healthClient := healthpb.NewHealthClient(conn)

		// Perform health check
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := healthClient.Check(ctx, &healthpb.HealthCheckRequest{})
		cancel()

		if err != nil {
			slog.Error("Health check failed", "service", service.exeFileName, "error", err)
			continue
		}

		if resp.Status != healthpb.HealthCheckResponse_SERVING {
			slog.Warn("Service unhealthy", "service", service.exeFileName, "status", resp.Status)
		}
	}
}
