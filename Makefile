BINARY_NAME=webanalyzer
BUILD_DIR=bin
CMD_PATH=./cmd/server
BIN=$(BUILD_DIR)/$(BINARY_NAME)

# COMPOSE to hold either docker-compose (default) or podman-compose
# Examples:
#   make docker-up                            - uses docker compose (default)
#   COMPOSE=podman-compose make docker-up     - uses podman-compose
#   export COMPOSE=podman-compose             - persists for session
COMPOSE ?= docker compose

.PHONY: all build run clean test fmt vet \
        docker-build docker-up docker-up-detach docker-down docker-logs docker-rebuild

all: build

build: clean
	go build -o $(BIN) $(CMD_PATH)

run: build
	$(BIN)

test:
	go test ./... -v -race

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)

docker-build:
	$(COMPOSE) build

docker-up:
	$(COMPOSE) up --build

docker-up-detach:
	$(COMPOSE) up --build -d

docker-down:
	$(COMPOSE) down

docker-logs:
	$(COMPOSE) logs -f

docker-rebuild:
	$(COMPOSE) down
	$(COMPOSE) up --build