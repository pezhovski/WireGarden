BINARY_NAME := server
OUTPUT_DIR := $(abspath ./build)
SERVER_MODULE := ./server

.PHONY: build
build:
	@mkdir -p $(OUTPUT_DIR)
	@pushd $(SERVER_MODULE) && go build -o $(OUTPUT_DIR)/$(BINARY_NAME) . && popd
	@chmod +x $(OUTPUT_DIR)/$(BINARY_NAME)
	@echo "Server build completed successfully"

.PHONY: clean
clean:
	@rm -rf $(OUTPUT_DIR)
	@echo "Build artifacts removed"

.PHONY: run
run:
	@go run $(SERVER_MODULE)

.PHONY: test
test:
	@go test ./...

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build   - Build the server binary"

.DEFAULT_GOAL := help
