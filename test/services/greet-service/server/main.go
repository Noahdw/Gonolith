package main

import (
	"context"
	"log"
	"net"

	pb "github.com/noahdw/Gonolith/test/greet-service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{
		Message: "Hello, " + req.Name + "!",
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:8088")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(s, healthServer)

	pb.RegisterGreeterServer(s, &server{})

	log.Println("Server starting on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
