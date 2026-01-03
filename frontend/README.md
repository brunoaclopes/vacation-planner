# Vacation Planner Frontend

A React-based single-page application for the Portuguese Vacation Planner. Features an interactive calendar, AI-powered chat assistant, and comprehensive settings management with full internationalization support.

## Technology Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| React | 18.2 | UI framework |
| TypeScript | 5.3 | Type safety |
| Vite | 5.0 | Build tool and dev server |
| Material UI | 5.15 | Component library |
| React Router | 6.21 | Client-side routing |
| Axios | 1.6 | HTTP client |
| date-fns | 3.2 | Date manipulation |

## Project Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── CalendarSummary.tsx    # Vacation days summary display
│   │   ├── ChatPanel.tsx          # AI chat interface
│   │   ├── HolidayNotification.tsx # Holiday loading status banner
│   │   ├── Layout.tsx             # App shell with navigation
│   │   ├── YearCalendar.tsx       # 12-month calendar grid
│   │   └── YearConfigPanel.tsx    # Year settings sidebar
│   ├── context/
│   │   ├── CalendarContext.tsx    # Global calendar state management
│   │   └── ThemeContext.tsx       # Dark/light theme provider
│   ├── i18n/
│   │   ├── I18nContext.tsx        # Internationalization provider
│   │   ├── index.ts               # i18n exports
│   │   ├── translations.ts        # Translation loader
│   │   ├── types.ts               # Language types
│   │   └── locales/
│   │       ├── en.ts              # English translations
│   │       ├── pt-PT.ts           # Portuguese translations
│   │       └── index.ts           # Locale exports
│   ├── pages/
│   │   ├── CalendarPage.tsx       # Main calendar view
│   │   └── SettingsPage.tsx       # Application settings
│   ├── services/
│   │   └── api.ts                 # Backend API client
│   ├── types/
│   │   └── index.ts               # TypeScript interfaces
│   ├── App.tsx                    # Root component with routing
│   ├── main.tsx                   # Application entry point
│   └── vite-env.d.ts              # Vite type declarations
├── public/                        # Static assets
├── Dockerfile                     # Multi-stage Docker build
├── nginx.conf                     # Production nginx configuration
├── package.json                   # Dependencies and scripts
├── tsconfig.json                  # TypeScript configuration
└── vite.config.ts                 # Vite configuration
```

## Features

### Interactive Calendar
- **12-month grid view** with color-coded days
- **Click to add/remove** vacation days
- **Visual indicators** for:
  - 🔴 Holidays (red)
  - 🔵 Manual vacations (blue)
  - 🟢 AI-optimized vacations (green)
  - ⚪ Weekends (gray)
- **Hover tooltips** with day details and holiday names

### Vacation Optimization
- **AI-powered optimization** suggests best vacation dates
- **Three strategies**:
  - Balanced (mix of long weekends and week blocks)
  - Long Weekends (maximize 3-4 day breaks)
  - Week Blocks (full weeks off)
- **Real-time suggestions** at bottom of calendar

### AI Chat Assistant
- **Context-aware** conversations about your calendar
- **Multi-model support** (GPT-4o, GPT-4o-mini, o1)
- **Language-aware** responses match UI language
- **Persistent history** per year

### Settings Management
- **AI Configuration**: Provider, model, API keys
- **Calendar Defaults**: Work week, vacation days
- **Location Settings**: City for municipal holidays
- **Language Selection**: English / Portuguese

### Internationalization
- **Full i18n support** for EN and PT-PT
- **Persistent language** preference (localStorage)
- **Dynamic translations** including holidays

## Components

### CalendarSummary
Displays vacation day statistics:
- Total vacation days available
- Days used (manual + optimized)
- Days remaining
- Reserved days

### ChatPanel
AI chat interface with:
- Message history display
- Input field with send button
- Model selector dropdown
- Clear history option
- Auto-scroll to latest message

### HolidayNotification
Banner showing holiday data status:
- Loading indicator
- Error messages
- Refresh button

### Layout
Application shell including:
- Top navigation bar
- Year selector (prev/next buttons)
- Theme toggle (dark/light)
- Language switcher
- Navigation links (Calendar, Settings)

### YearCalendar
Main calendar grid:
- 12 months in 4x3 grid (responsive)
- Day cells with click handlers
- Color coding by day type
- Holiday/vacation tooltips

### YearConfigPanel
Sidebar configuration:
- Vacation days input
- Reserved days input
- Work week checkboxes
- Optimization strategy selector
- Optimizer notes textarea
- Optimize button
- Clear optimized button

## Pages

### CalendarPage (`/`)
Main view combining:
- YearCalendar (center)
- YearConfigPanel (right sidebar)
- CalendarSummary (top)
- ChatPanel (expandable bottom)
- AI suggestion banner

### SettingsPage (`/settings`)
Configuration sections:
- **AI Assistant**: Provider, model, API key
- **Backend**: Port configuration
- **Calendar**: Default vacation days, work week
- **Location**: Work city (for municipal holidays)
- **Dictionary**: Calendarific API key
- **Language**: EN/PT-PT selector

## State Management

### CalendarContext
Global state provider managing:
```typescript
interface CalendarContextType {
  year: number;
  setYear: (year: number) => void;
  calendar: CalendarResponse | null;
  loading: boolean;
  error: string | null;
  refreshCalendar: () => Promise<void>;
  addVacation: (date: string, note?: string) => Promise<void>;
  removeVacation: (date: string) => Promise<void>;
  optimizeVacations: () => Promise<void>;
  clearOptimizedVacations: () => Promise<void>;
  updateYearConfig: (config: Partial<YearConfig>) => Promise<void>;
}
```

### ThemeContext
Theme management:
```typescript
interface ThemeContextType {
  mode: 'light' | 'dark';
  toggleTheme: () => void;
}
```

### I18nContext
Internationalization:
```typescript
interface I18nContextType {
  language: Language;           // 'en' | 'pt-PT'
  setLanguage: (lang: Language) => void;
  t: Translations;              // Translation object
}
```

## TypeScript Interfaces

### CalendarResponse
```typescript
interface CalendarResponse {
  year: number;
  config: YearConfig;
  days: CalendarDay[];
  holidays: Holiday[];
  vacations: VacationDay[];
  optimal_vacations: VacationDay[];
  summary: CalendarSummary;
}
```

### CalendarDay
```typescript
interface CalendarDay {
  date: string;
  day_of_week: string;
  is_weekend: boolean;
  is_holiday: boolean;
  holiday_name?: string;
  is_vacation: boolean;
  is_optimal: boolean;
  is_manual: boolean;
  note?: string;
}
```

### YearConfig
```typescript
interface YearConfig {
  year: number;
  vacation_days: number;
  reserved_days: number;
  optimization_strategy: 'balanced' | 'long_weekends' | 'week_blocks';
  work_week: string[];
  optimizer_notes: string;
}
```

## API Integration

All API calls are centralized in `src/services/api.ts`:

```typescript
// Calendar
getCalendar(year: number): Promise<CalendarResponse>
optimizeVacations(year: number): Promise<OptimizationResult>
clearOptimizedVacations(year: number): Promise<void>
getVacationSuggestions(year: number, language: string): Promise<Suggestion>

// Vacations
getVacations(year: number): Promise<VacationDay[]>
addVacation(year: number, date: string, note?: string): Promise<VacationDay>
removeVacation(year: number, date: string): Promise<void>
bulkUpdateVacations(year: number, dates: string[]): Promise<void>

// Holidays
getHolidays(year: number): Promise<Holiday[]>
refreshHolidays(year: number): Promise<void>
getAvailableCities(): Promise<City[]>

// Configuration
getYearConfig(year: number): Promise<YearConfig>
updateYearConfig(year: number, config: Partial<YearConfig>): Promise<YearConfig>
copyYearConfig(year: number, sourceYear: number): Promise<YearConfig>

// Settings
getSettings(): Promise<Settings>
updateSettings(settings: Partial<Settings>): Promise<Settings>
getSetting(key: string): Promise<string>
updateSetting(key: string, value: string): Promise<void>

// Chat
getAvailableModels(): Promise<Model[]>
chat(year: number, message: string): Promise<ChatResponse>
getChatHistory(year: number): Promise<ChatMessage[]>
clearChatHistory(year: number): Promise<void>
```

## Translations

### Adding a New Language

1. Create locale file `src/i18n/locales/xx.ts`:
```typescript
import { Translations } from '../types';

export const xx: Translations = {
  // Copy structure from en.ts
  common: { ... },
  calendar: { ... },
  settings: { ... },
  chat: { ... },
};
```

2. Add to `src/i18n/locales/index.ts`:
```typescript
export { xx } from './xx';
```

3. Update `src/i18n/types.ts`:
```typescript
export type Language = 'en' | 'pt-PT' | 'xx';
```

4. Update `src/i18n/translations.ts`:
```typescript
import { xx } from './locales';
export const translations: Record<Language, Translations> = {
  'en': en,
  'pt-PT': ptPT,
  'xx': xx,
};
```

## Styling

### Theme Configuration
Material UI theme with:
- Light/dark mode support
- Portuguese-inspired color palette
- Responsive breakpoints
- Custom component overrides

### Color Coding (Calendar)
| Day Type | Light Mode | Dark Mode |
|----------|------------|-----------|
| Holiday | `#ef5350` | `#c62828` |
| Manual Vacation | `#42a5f5` | `#1565c0` |
| Optimized Vacation | `#66bb6a` | `#2e7d32` |
| Weekend | `#e0e0e0` | `#424242` |
| Today | `#fff59d` | `#f9a825` |

## Running Locally

### Prerequisites
- Node.js 18+
- npm or yarn

### Development
```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

Development server starts at `http://localhost:5173`

### Building
```bash
# Type check and build
npm run build

# Preview production build
npm run preview
```

### Linting
```bash
npm run lint
```

## Docker

### Build
```bash
docker build -t vacation-planner-frontend .
```

### Run
```bash
docker run -p 8080:80 vacation-planner-frontend
```

The Dockerfile uses multi-stage build:
1. **Builder stage**: Node.js builds the React app
2. **Runtime stage**: Nginx serves static files and proxies API

### Nginx Configuration
- Serves static files from `/usr/share/nginx/html`
- SPA routing (all routes → `index.html`)
- API proxy: `/api/*` → `backend:8080`
- Gzip compression enabled

## Environment Configuration

### Development
API calls proxy through Vite to `http://localhost:8080`

### Production (Docker)
Nginx proxies `/api/*` to the backend container

## Browser Support

- Chrome 90+
- Firefox 90+
- Safari 14+
- Edge 90+

## Performance Optimizations

- **Code splitting** via React Router lazy loading
- **Memoization** of expensive calendar calculations
- **Debounced** API calls for config updates
- **Cached** calendar data with smart invalidation
- **Optimistic UI** updates for vacation toggles

## Accessibility

- **Keyboard navigation** for calendar
- **ARIA labels** on interactive elements
- **Color contrast** meets WCAG AA
- **Focus indicators** visible
- **Screen reader** friendly structure

## License

MIT
