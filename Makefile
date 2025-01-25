BINARY_NAME=go-w
DOCKER_IMAGE_NAME=go-w
DOCKER_TAG=latest

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .

install:
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	sudo mv $(BINARY_NAME) /usr/local/bin/

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)

run:
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) .

docker-run:
	@echo "Running Docker container..."
	docker run --rm $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

docker-push:
	@echo "Pushing Docker image to registry..."
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

docker-test:
	@echo "Running tests inside Docker container..."
	docker build -t $(DOCKER_IMAGE_NAME)-test:$(DOCKER_TAG) --target=builder .
	docker run --rm $(DOCKER_IMAGE_NAME)-test:$(DOCKER_TAG) go test -v ./...

docker-clean:
	@echo "Cleaning up Docker artifacts..."
	docker rmi -f $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) || true
	docker rmi -f $(DOCKER_IMAGE_NAME)-test:$(DOCKER_TAG) || true

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
	@echo "  docker-test  - Run tests inside Docker container"
	@echo "  docker-clean - Clean up Docker artifacts"
	@echo "  help         - Show this help message"

.PHONY: all build install test clean run docker-build docker-run docker-push docker-test docker-clean help