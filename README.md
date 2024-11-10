# README

## 
# Construir imágenes
```bash
docker build -f Dockerfile.master -t dev.rubentxu/master:latest .
docker build -f Dockerfile.agent -t dev.rubentxu/agent:latest .
```

# Publicar imágenes
```bash
docker push your-registry/master:latest
docker push your-registry/agent:latest
```

```yaml
# k8s/master-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: master-server
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: master
        image: your-registry/master:latest
        ports:
        - containerPort: 50051
---
# k8s/agent-daemonset.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: agent
spec:
  template:
    spec:
      containers:
      - name: agent
        image: your-registry/agent:latest
        env:
        - name: MASTER_ADDRESS
          value: "master-server:50051"
        - name: AGENT_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
```

---

Te explico el proceso completo de cómo crear y comunicar los dos binarios.

1. Primero, vamos a organizar la estructura del proyecto para tener dos binarios separados:

```
devops-platform/
├── cmd/
│   ├── agent/              # Binario del agente
│   │   └── main.go
│   └── master/            # Binario del servidor maestro
│       └── main.go
├── internal/
│   ├── proto/            # Definiciones compartidas de gRPC
│   │   └── agent.proto
│   ├── shared/           # Código compartido entre agent y master
│   │   └── version/
│   │       └── version.go
│   ├── master/          # Código específico del servidor
│   │   └── server/
│   │       └── grpc.go
│   └── agent/           # Código específico del agente
│       └── client/
│           └── grpc.go
├── Makefile             # Para automatizar la compilación
└── go.mod
```

2. Definición del protocolo gRPC (el contrato entre cliente y servidor):

```protobuf
// internal/proto/agent.proto
syntax = "proto3";

package agent;

option go_package = "devops-platform/internal/proto/agent";

service AgentService {
    rpc Connect(stream AgentMessage) returns (stream MasterMessage) {}
}

message AgentMessage {
    string agent_id = 1;
    oneof content {
        AgentRegistration registration = 2;
        StatusUpdate status_update = 3;
        Heartbeat heartbeat = 4;
    }
}

message MasterMessage {
    string message_id = 1;
    oneof content {
        Command command = 2;
        Acknowledgement ack = 3;
    }
}

// ... (resto de las definiciones de mensajes)
```

3. Makefile para compilar ambos binarios y generar el código gRPC:

```makefile
.PHONY: all proto build-master build-agent clean

# Variables
BINARY_DIR=bin
PROTO_DIR=internal/proto
MASTER_BINARY=$(BINARY_DIR)/master
AGENT_BINARY=$(BINARY_DIR)/agent

all: proto build-master build-agent

# Generar código desde .proto
proto:
    protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        $(PROTO_DIR)/agent.proto

# Compilar el servidor maestro
build-master:
    go build -o $(MASTER_BINARY) cmd/master/main.go

# Compilar el agente
build-agent:
    go build -o $(AGENT_BINARY) cmd/agent/main.go

clean:
    rm -rf $(BINARY_DIR)
```

4. Servidor Maestro:

```go
// cmd/master/main.go
package main

import (
    "devops-platform/internal/master/server"
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

// internal/master/server/grpc.go
package server

import (
    pb "devops-platform/internal/proto/agent"
    "google.golang.org/grpc"
    "sync"
)

type GRPCServer struct {
    pb.UnimplementedAgentServiceServer
    agents map[string]*AgentConnection
    mu     sync.RWMutex
}

func NewGRPCServer() *grpc.Server {
    server := &GRPCServer{
        agents: make(map[string]*AgentConnection),
    }
    
    grpcServer := grpc.NewServer()
    pb.RegisterAgentServiceServer(grpcServer, server)
    return grpcServer
}

// Implementación del método Connect definido en el proto
func (s *GRPCServer) Connect(stream pb.AgentService_ConnectServer) error {
    // Manejar la conexión del agente
    // ...
    return nil
}
```

5. Cliente Agente:

```go
// cmd/agent/main.go
package main

import (
    "context"
    "devops-platform/internal/agent/client"
    "log"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    // Crear el cliente gRPC
    agentClient, err := client.NewGRPCClient(
        "localhost:50051",  // Dirección del servidor maestro
        os.Getenv("AGENT_ID"),
    )
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    // Manejar señales de término
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

    // Conectar al servidor
    if err := agentClient.Connect(ctx); err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }

    // Esperar señal de término
    <-sigCh
    log.Println("Shutting down agent...")
    agentClient.Disconnect()
}

// internal/agent/client/grpc.go
package client

import (
    "context"
    pb "devops-platform/internal/proto/agent"
    "google.golang.org/grpc"
)

type GRPCClient struct {
    conn    *grpc.ClientConn
    stream  pb.AgentService_ConnectClient
    agentID string
}

func NewGRPCClient(serverAddr, agentID string) (*GRPCClient, error) {
    return &GRPCClient{
        agentID: agentID,
    }, nil
}

func (c *GRPCClient) Connect(ctx context.Context) error {
    // Establecer conexión gRPC
    conn, err := grpc.Dial(c.serverAddr, grpc.WithInsecure())
    if err != nil {
        return err
    }
    c.conn = conn

    // Crear el cliente gRPC
    client := pb.NewAgentServiceClient(conn)
    
    // Iniciar el stream bidireccional
    stream, err := client.Connect(ctx)
    if err != nil {
        return err
    }
    c.stream = stream

    // Iniciar rutinas de manejo de mensajes
    go c.handleServerMessages(ctx)
    go c.sendHeartbeats(ctx)

    return nil
}

func (c *GRPCClient) Disconnect() {
    if c.conn != nil {
        c.conn.Close()
    }
}
```

6. Para compilar y ejecutar:

```bash
# Generar código gRPC y compilar ambos binarios
make all

# Ejecutar el servidor maestro
./bin/master

# En otra terminal, ejecutar el agente
AGENT_ID=agent-1 ./bin/agent
```

El flujo de comunicación funciona así:

1. El servidor maestro inicia y escucha en el puerto 50051 (o el que configures)

2. Cuando un agente se inicia:
   - Establece una conexión gRPC con el servidor
   - Crea un stream bidireccional
   - Envía un mensaje de registro

3. La comunicación es bidireccional:
   - El servidor puede enviar comandos al agente
   - El agente puede enviar actualizaciones de estado
   - El agente envía heartbeats periódicos

4. Ejemplo de compilación y ejecución en Kubernetes:

```yaml
# k8s/master-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: master-server
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: master
        image: your-registry/master:latest
        ports:
        - containerPort: 50051
---
# k8s/agent-daemonset.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: agent
spec:
  template:
    spec:
      containers:
      - name: agent
        image: your-registry/agent:latest
        env:
        - name: MASTER_ADDRESS
          value: "master-server:50051"
        - name: AGENT_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
```

Para construir las imágenes Docker:

```dockerfile
# Dockerfile.master
FROM golang:1.21 as builder
WORKDIR /app
COPY . .
RUN make build-master

FROM alpine:3.18
COPY --from=builder /app/bin/master /usr/local/bin/
CMD ["master"]

# Dockerfile.agent
FROM golang:1.21 as builder
WORKDIR /app
COPY . .
RUN make build-agent

FROM alpine:3.18
COPY --from=builder /app/bin/agent /usr/local/bin/
CMD ["agent"]
```

Para construir y publicar las imágenes:

```bash
# Construir imágenes
docker build -f Dockerfile.master -t your-registry/master:latest .
docker build -f Dockerfile.agent -t your-registry/agent:latest .

# Publicar imágenes
docker push your-registry/master:latest
docker push your-registry/agent:latest
```

Esta estructura permite:

1. Separación clara entre código del servidor y del agente
2. Comunicación bidireccional a través de gRPC
3. Fácil compilación y despliegue
4. Base para agregar más funcionalidades

La comunicación gRPC proporciona:

1. Streaming bidireccional
2. Serialización eficiente
3. Tipado fuerte
4. Generación automática de código
5. Manejo de errores robusto