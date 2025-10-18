.PHONY: build test run docker-up docker-down docker-db-up docker-db-stop docker-restart-app docker-logs lint migrate-up migrate-down migrate-create

build:
	@echo "Building application..."
	go build -buildvcs=false -o shortener cmd/shortener/main.go
	@echo "Build complete: shortener"

test:
	@echo "Running unit tests..."
	go test ./...

run:
	@echo "Running application..."
	go run cmd/shortener/main.go

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

docker-db-up:
	@echo "Starting database and running migrations..."
	docker-compose up -d postgres
	docker-compose up migrate

docker-db-stop:
	@echo "Stopping database..."
	docker-compose stop postgres

docker-restart-app:
	@echo "Restarting application..."
	docker-compose restart app

lint:
	@echo "Running linter..."
	go vet ./...
	@echo "Running static analysis..."
	go run cmd/staticlint/main.go ./...

migrate-up:
	@echo "Running database migrations..."
	migrate -path ./migrations -database 'postgres://postgres:shortener@localhost:5432/shortener?sslmode=disable' up

migrate-down:
	@echo "Rolling back database migrations..."
	migrate -path ./migrations -database 'postgres://postgres:shortener@localhost:5432/shortener?sslmode=disable' down
