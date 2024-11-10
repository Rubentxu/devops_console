.PHONY: all proto check-proto-tools build-master build-agent clean

# Variables de directorios
PROTO_DIR := internal/infrastructure/agent/proto/agent/v1
GO_OUT_DIR := ./
BINARY_DIR := bin
AGENT_BINARY := $(BINARY_DIR)/agent
MASTER_BINARY := $(BINARY_DIR)/master

# Verificar herramientas necesarias
.PHONY: check-proto-tools
check-proto-tools:
	@which protoc > /dev/null || (echo "Error: protoc no está instalado" && exit 1)
	@which protoc-gen-go > /dev/null || (echo "Error: protoc-gen-go no está instalado" && exit 1)
	@which protoc-gen-go-grpc > /dev/null || (echo "Error: protoc-gen-go-grpc no está instalado" && exit 1)

# Generar código desde .proto
.PHONY: proto
proto: check-proto-tools
	protoc --go_out=$(GO_OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto

# Limpiar archivos generados
.PHONY: clean-proto
clean-proto:
	find $(GO_OUT_DIR) -name "*.pb.go" -delete

# Compilar el servidor maestro
build-master:
	go build -o $(MASTER_BINARY) cmd/master/main.go

# Compilar el agente
build-agent:
	mkdir -p $(BINARY_DIR)
	go build -o $(AGENT_BINARY) cmd/agent/main.go

clean:
	rm -rf $(BINARY_DIR)