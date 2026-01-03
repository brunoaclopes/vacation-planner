# Vacation Planner Backend

A Go-based REST API backend for the Portuguese Vacation Planner application. It provides vacation optimization, holiday management, AI-powered chat assistance, and calendar data persistence.

## Technology Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| Go | 1.21+ | Core language |
| Gin | 1.9.1 | HTTP web framework |
| SQLite3 | - | Embedded database |
| go-openai | 1.17.9 | OpenAI/GitHub Models API client |
| gin-cors | 1.5.0 | CORS middleware |

## Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── handlers.go      # Core API handlers (calendar, vacations, settings)
│   │   │   └── chat.go          # AI chat handlers
│   │   └── server.go            # HTTP server setup and routing
│   ├── database/
│   │   └── database.go          # SQLite initialization and schema
│   ├── holidays/
│   │   ├── portuguese.go        # Portuguese holiday calculations (Easter-based)
│   │   └── service.go           # Holiday service with Calendarific API support
│   ├── models/
│   │   └── models.go            # Data models and types
│   └── optimizer/
│       └── optimizer.go         # Vacation optimization algorithms
├── Dockerfile                   # Multi-stage Docker build
├── go.mod                       # Go module definition
└── go.sum                       # Dependency checksums
```

## API Endpoints

### Health Check
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check, returns `{"status": "ok"}` |

### Calendar
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/calendar/:year` | Get full calendar with holidays, vacations, and summary |
| POST | `/api/calendar/:year/optimize` | Run vacation optimization algorithm |
| DELETE | `/api/calendar/:year/optimized` | Clear AI-optimized vacation days |
| GET | `/api/calendar/:year/suggestions` | Get AI-powered vacation suggestions |

### Vacations
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/vacations/:year` | Get all manual vacation days for a year |
| POST | `/api/vacations/:year` | Add a vacation day |
| DELETE | `/api/vacations/:year/:date` | Remove a vacation day |
| PUT | `/api/vacations/:year/bulk` | Bulk update vacation days |

### Holidays
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/holidays/:year` | Get all holidays for a year |
| GET | `/api/holidays/:year/status` | Get holiday loading status |
| GET | `/api/holidays/status` | Get all years' holiday statuses |
| POST | `/api/holidays/:year/refresh` | Refresh holidays from external API |
| GET | `/api/cities` | Get available Portuguese cities for municipal holidays |

### Year Configuration
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/config/:year` | Get year configuration |
| PUT | `/api/config/:year` | Update year configuration |
| POST | `/api/config/:year/copy-from/:sourceYear` | Copy configuration from another year |

### Settings
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/settings` | Get all application settings |
| PUT | `/api/settings` | Update multiple settings |
| GET | `/api/settings/:key` | Get a specific setting |
| PUT | `/api/settings/:key` | Update a specific setting |

### AI Chat
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/models` | Get available AI models |
| POST | `/api/chat/:year` | Send chat message to AI assistant |
| GET | `/api/chat/:year/history` | Get chat history for a year |
| DELETE | `/api/chat/:year/history` | Clear chat history |

### Presets
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/presets/work-week` | Get work week preset options |
| GET | `/api/presets/strategies` | Get optimization strategy options |

## Data Models

### YearConfig
```go
type YearConfig struct {
    Year                 int      `json:"year"`
    VacationDays         int      `json:"vacation_days"`          // Total vacation days available
    ReservedDays         int      `json:"reserved_days"`          // Days reserved (not for optimization)
    OptimizationStrategy string   `json:"optimization_strategy"`  // "balanced", "long_weekends", "week_blocks"
    WorkWeek             []string `json:"work_week"`              // e.g., ["monday","tuesday","wednesday","thursday","friday"]
    OptimizerNotes       string   `json:"optimizer_notes"`        // Custom notes for AI optimizer
}
```

### VacationDay
```go
type VacationDay struct {
    Year     int    `json:"year"`
    Date     string `json:"date"`      // Format: "2026-01-15"
    IsManual bool   `json:"is_manual"` // true for user-added, false for AI-optimized
    Note     string `json:"note"`      // Optional note
}
```

### Holiday
```go
type Holiday struct {
    Year int    `json:"year"`
    Date string `json:"date"`
    Name string `json:"name"`
    Type string `json:"type"` // "national", "municipal", "optional"
}
```

### CalendarDay
```go
type CalendarDay struct {
    Date        string `json:"date"`
    DayOfWeek   string `json:"day_of_week"`
    IsWeekend   bool   `json:"is_weekend"`
    IsHoliday   bool   `json:"is_holiday"`
    HolidayName string `json:"holiday_name,omitempty"`
    IsVacation  bool   `json:"is_vacation"`
    IsOptimal   bool   `json:"is_optimal"`    // AI-suggested vacation
    IsManual    bool   `json:"is_manual"`     // User-added vacation
    Note        string `json:"note,omitempty"`
}
```

## Database Schema

SQLite database with the following tables:

```sql
-- Application settings (API keys, defaults)
CREATE TABLE settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT NOT NULL UNIQUE,
    value TEXT NOT NULL
);

-- Year-specific configuration
CREATE TABLE year_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    year INTEGER NOT NULL UNIQUE,
    vacation_days INTEGER DEFAULT 22,
    reserved_days INTEGER DEFAULT 0,
    optimization_strategy TEXT DEFAULT 'balanced',
    work_week TEXT DEFAULT '["monday","tuesday","wednesday","thursday","friday"]',
    optimizer_notes TEXT DEFAULT ''
);

-- Manual vacation days
CREATE TABLE vacation_days (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    year INTEGER NOT NULL,
    date TEXT NOT NULL,
    is_manual INTEGER DEFAULT 1,
    note TEXT,
    UNIQUE(year, date)
);

-- AI-optimized vacation days
CREATE TABLE optimal_vacations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    year INTEGER NOT NULL,
    date TEXT NOT NULL,
    block_id INTEGER,
    consecutive_days INTEGER,
    UNIQUE(year, date)
);

-- Cached holidays
CREATE TABLE holidays (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    year INTEGER NOT NULL,
    date TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT DEFAULT 'national',
    location TEXT DEFAULT '',
    UNIQUE(year, date, type, location)
);

-- AI chat history
CREATE TABLE chat_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    year INTEGER NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);
```

## Optimization Strategies

The optimizer supports three strategies:

| Strategy | Description |
|----------|-------------|
| `balanced` | Mix of long weekends and week-long blocks |
| `long_weekends` | Prioritize extending weekends (3-4 day breaks) |
| `week_blocks` | Prioritize full week vacations (7+ consecutive days) |

The algorithm considers:
- Public holidays and their proximity to weekends
- Work week configuration (supports 4-day weeks, custom schedules)
- Reserved days that shouldn't be auto-optimized
- Maximum consecutive days off vs vacation days used ratio

## AI Integration

Supports multiple AI providers:

| Provider | Models | Configuration |
|----------|--------|---------------|
| GitHub Models | GPT-4o, GPT-4o-mini, o1, o1-mini | `ai_provider: "github"`, uses GitHub token |
| OpenAI | GPT-4, GPT-3.5-turbo | `ai_provider: "openai"`, requires API key |

The AI assistant can:
- Suggest optimal vacation periods based on calendar
- Answer questions about Portuguese holidays
- Provide vacation planning advice
- Respond in the UI's selected language (EN/PT-PT)

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `GIN_MODE` | `debug` | Gin mode (`debug`, `release`) |
| `PORT` | `8080` | Server port |

Settings stored in database:
- `openai_api_key` - OpenAI API key
- `ai_provider` - AI provider (`github` or `openai`)
- `ai_model` - AI model to use
- `work_city` - City for municipal holidays
- `calendarific_api_key` - External holiday API key

## Running Locally

### Prerequisites
- Go 1.21+
- GCC (for SQLite CGO compilation)

### Development
```bash
cd backend

# Install dependencies
go mod download

# Run server
go run cmd/server/main.go

# Or build and run
go build -o server cmd/server/main.go
./server
```

Server starts at `http://localhost:8080`

### Building for Production
```bash
CGO_ENABLED=1 go build -a -ldflags '-linkmode external -extldflags "-static"' -o server cmd/server/main.go
```

## Docker

### Build
```bash
docker build -t vacation-planner-backend .
```

### Run
```bash
docker run -p 8080:8080 -v vacation-data:/app/data vacation-planner-backend
```

The Dockerfile uses multi-stage build:
1. **Builder stage**: Compiles Go with CGO for SQLite
2. **Runtime stage**: Minimal Alpine image with the binary

## Portuguese Holidays

The backend includes a comprehensive Portuguese holiday calculator:

### Fixed Holidays
- New Year's Day (January 1)
- Freedom Day (April 25)
- Labour Day (May 1)
- Portugal Day (June 10)
- Assumption of Mary (August 15)
- Republic Day (October 5)
- All Saints' Day (November 1)
- Restoration of Independence (December 1)
- Immaculate Conception (December 8)
- Christmas Day (December 25)

### Easter-based Holidays (calculated annually)
- Carnival (47 days before Easter) - Optional
- Good Friday (2 days before Easter)
- Easter Sunday
- Corpus Christi (60 days after Easter)

### Municipal Holidays
Supports city-specific holidays for all Portuguese municipalities (e.g., Lisbon - June 13, Porto - June 24).

## License

MIT
