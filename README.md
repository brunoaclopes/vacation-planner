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

## Quick Start with Docker

### Using Pre-built Image (Recommended)

```bash
docker run -d \
  --name vacation-planner \
  -p 8080:80 \
  -v vacation-planner-data:/app/data \
  ghcr.io/brunoaclopes/vacation-planner:latest
```

The app will be available at http://localhost:8080

### TrueNAS SCALE Deployment

1. Go to **Apps** → **Discover Apps** → **Custom App**
2. Configure:
   - **Image**: `ghcr.io/brunoaclopes/vacation-planner:latest`
   - **Port**: Map container port `80` to your desired host port (e.g., `8080`)
   - **Storage**: Mount a host path or volume to `/app/data` for persistence
3. Deploy and access at `http://TRUENAS_IP:8080`

### Using Docker Compose

```bash
# Clone the repo
git clone https://github.com/brunoaclopes/vacation-planner.git
cd vacation-planner

# Build and start containers
docker compose up -d

# View logs
docker compose logs -f

# Stop containers
docker compose down
```

## Development Setup

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
- Frontend: 5173 (dev) / 80 (Docker)

Data is stored in SQLite at `/app/data/vacation_planner.db`.

## Docker Images

| Image | Description |
|-------|-------------|
| `ghcr.io/brunoaclopes/vacation-planner:latest` | All-in-one (recommended) |
| `ghcr.io/brunoaclopes/vacation-planner-backend:latest` | Backend only |
| `ghcr.io/brunoaclopes/vacation-planner-frontend:latest` | Frontend only |

## License

MIT
