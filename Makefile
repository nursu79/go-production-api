.PHONY: build run up down clean test tidy

APP_NAME = go-production-api
MAIN_FILE = cmd/api/main.go

# Load basic .env file for local 'make run'
include .env.example
export $(shell sed 's/=.*//' .env.example)

build:
	@echo "Building application..."
	go build -o bin/$(APP_NAME) $(MAIN_FILE)

run: build
	@echo "Running locally..."
	./bin/$(APP_NAME)

up:
	@echo "Starting services with Docker Compose..."
	docker compose up --build -d

down:
	@echo "Stopping Docker Compose services..."
	docker compose down

clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	docker compose down -v
	docker system prune -f

test:
	go test -v ./...

tidy:
	go mod tidy

migrate-up:
	@echo "Running up migrations..."
	migrate -path migrations -database "$(DB_URL)" -verbose up

migrate-down:
	@echo "Running down migrations..."
	migrate -path migrations -database "$(DB_URL)" -verbose down -all
