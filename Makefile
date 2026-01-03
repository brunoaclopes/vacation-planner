.PHONY: all backend frontend install dev clean

all: install dev

# Install dependencies
install:
	@echo "Installing backend dependencies..."
	cd backend && go mod tidy
	@echo "Installing frontend dependencies..."
	cd frontend && npm install

# Run development servers
dev:
	@echo "Starting development servers..."
	@make -j2 backend frontend

backend:
	@echo "Starting Go backend on port 8080..."
	cd backend && go run cmd/server/main.go

frontend:
	@echo "Starting React frontend on port 5173..."
	cd frontend && npm run dev

# Build for production
build:
	@echo "Building backend..."
	cd backend && go build -o ../dist/server cmd/server/main.go
	@echo "Building frontend..."
	cd frontend && npm run build
	@echo "Copying frontend build..."
	cp -r frontend/dist dist/frontend

# Clean build artifacts
clean:
	rm -rf dist/
	rm -rf frontend/node_modules
	rm -rf backend/data/

# Run tests
test:
	cd backend && go test ./...
	cd frontend && npm test

# Docker operations
docker-build:
	@echo "Building Docker images..."
	docker compose build

docker-up:
	@echo "Starting Docker containers..."
	docker compose up -d

docker-down:
	@echo "Stopping Docker containers..."
	docker compose down

docker-logs:
	docker compose logs -f

docker-clean:
	@echo "Removing Docker containers, images, and volumes..."
	docker compose down -v --rmi local

# Database operations
db-reset:
	rm -f backend/data/calendar.db
	@echo "Database reset. Will be recreated on next run."
