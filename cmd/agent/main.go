package main

import (
	"context"
	"devops_console/internal/infrastructure/agent"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Configuración del agente
	agentID := "agent-1"
	masterAddr := "localhost:50051"
	creds := insecure.NewCredentials()

	// Crear una nueva instancia del agente
	agent := agent.NewAgent(agentID, masterAddr, creds)

	// Crear un contexto con cancelación
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Manejar señales del sistema para una salida limpia
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Recibida señal de terminación, cerrando agente...")
		cancel()
	}()

	// Iniciar el agente
	if err := agent.Start(ctx); err != nil {
		log.Fatalf("Error al iniciar el agente: %v", err)
	}

	// Esperar a que el contexto se cancele
	<-ctx.Done()
	log.Println("Agente cerrado correctamente")
}
