.PHONY: all build up down clean sandbox-images

all: sandbox-images build up

# Build sandbox images (Go + C++ runtimes) separately first
sandbox-images:
	docker compose build sandbox-go sandbox-cpp

# Build all service images
build: sandbox-images
	docker compose build

# Start all services
up:
	docker compose up -d

# Stop all services
down:
	docker compose down

# Full cleanup
clean: down
	docker compose down -v --rmi all
