# Makefile for go-w

# Variables
BINARY_NAME=go-w
DOCKER_IMAGE_NAME=go-w
DOCKER_TAG=latest

# Default target
all: build

# Build the project
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .

# Install the binary to /usr/local/bin
install:
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo mv $(BINARY_NAME) /usr/local/bin/

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Build the Docker image
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) .

# Run the Docker container
docker-run:
	@echo "Running Docker container..."
	docker run --rm $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

# Push the Docker image to a registry
docker-push:
	@echo "Pushing Docker image to registry..."
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

# Help target
help:
	@echo "Available targets:"
	@echo "  build        - Build the project"
	@echo "  install      - Install the binary to /usr/local/bin"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean up build artifacts"
	@echo "  run          - Run the application"
	@echo "  docker-build - Build the Docker image"
	@echo "  docker-run   - Run the Docker container"
	@echo "  docker-push  - Push the Docker image to a registry"
	@echo "  help         - Show this help message"

.PHONY: all build install test clean run docker-build docker-run docker-push help