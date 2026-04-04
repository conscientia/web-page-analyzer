BINARY_NAME=webanalyzer
BUILD_DIR=bin
CMD_PATH=./cmd/server
BIN= $(BUILD_DIR)/$(BINARY_NAME)

.PHONY: all build run clean test

all: build

build: clean
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)

test:
	go test ./... -v

clean:
	rm -rf $(BUILD_DIR)

fmt:
	go fmt ./...

vet:
	go vet ./...

run: build
	$(BIN)