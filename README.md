# Portuguese Vacation Planner

A monorepo application to optimize your vacation days around Portuguese holidays.

## Structure

- `backend/` - Go API server with SQLite database
- `frontend/` - React TypeScript app with Material UI

## Features

- 📅 Year calendar view with Portuguese holidays
- 🧮 Vacation optimization algorithms
- 💬 AI-powered chat to interact with calendar
- ⚙️ Configurable settings (work week, ports, API keys)
- 📊 Multiple year support with independent configurations
- 🌍 Internationalization (English and Portuguese)

## Getting Started

### Using Docker (Recommended)

The easiest way to run the application is using Docker:

```bash
# Build and start containers
docker compose up -d

# View logs
docker compose logs -f

# Stop containers
docker compose down
```

The app will be available at http://localhost:8080

### Using Make

```bash
# Install dependencies and start dev servers
make all

# Or separately:
make install  # Install dependencies
make dev      # Start both servers
```

### Manual Setup

#### Backend

```bash
cd backend
go mod download
go run cmd/server/main.go
```

#### Frontend

```bash
cd frontend
npm install
npm run dev
```

## Docker Commands

```bash
make docker-build   # Build Docker images
make docker-up      # Start containers
make docker-down    # Stop containers
make docker-logs    # View logs
make docker-clean   # Remove containers, images, and volumes
```

## Configuration

Default ports:
- Backend: 8080
- Frontend: 5173 (dev) / 8080 (Docker)

These can be changed in the settings page.
