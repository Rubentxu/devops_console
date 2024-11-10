package main

import (
	"devops_console/internal/infrastructure/orchestrator/server"
	"log"
	"net"
)

func main() {
	// Crear el listener TCP
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Crear e iniciar el servidor gRPC
	grpcServer := server.NewGRPCServer()
	log.Printf("Starting gRPC server on %s", lis.Addr().String())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
