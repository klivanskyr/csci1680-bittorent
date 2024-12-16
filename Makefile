# Makefile

.PHONY: all client server clean

# Directories
CLIENT_DIR := cmd/client
SERVER_DIR := cmd/server

# Binaries
CLIENT_BIN := bin/client
SERVER_BIN := bin/server

# Build all binaries
all: client server

# Build client binary
client:
	@echo "Building client..."
	@cd $(CLIENT_DIR) && wails build -o ../../../../$(CLIENT_BIN)

# Build server binary
server:
	@echo "Building server..."
	@go build -o $(SERVER_BIN) $(SERVER_DIR)/main.go

# Run client
run-client: client
	@echo "Running client..."
	@$(CLIENT_BIN)

# Run server
run-server: server
	@echo "Running server..."
	@$(SERVER_BIN)

# Clean up binaries
clean:
	@echo "Cleaning up..."
	@rm -f $(CLIENT_BIN) $(SERVER_BIN)