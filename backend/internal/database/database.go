package database

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Initialize creates a SQLite database connection
func Initialize(dbPath string) (*sql.DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := createTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	schema := `
	-- Settings table for global and year-specific settings
	CREATE TABLE IF NOT EXISTS settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL UNIQUE,
		value TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Year configurations
	CREATE TABLE IF NOT EXISTS year_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL UNIQUE,
		vacation_days INTEGER DEFAULT 22,
		reserved_days INTEGER DEFAULT 0,
		optimization_strategy TEXT DEFAULT 'balanced',
		work_week TEXT DEFAULT '["monday","tuesday","wednesday","thursday","friday"]',
		optimizer_notes TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Manual vacation days set by user
	CREATE TABLE IF NOT EXISTS vacation_days (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL,
		date TEXT NOT NULL,
		is_manual BOOLEAN DEFAULT TRUE,
		note TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(year, date)
	);

	-- Calculated optimal vacation days
	CREATE TABLE IF NOT EXISTS optimal_vacations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL,
		date TEXT NOT NULL,
		block_id INTEGER,
		consecutive_days INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(year, date)
	);

	-- Portuguese holidays (can vary by year for some)
	CREATE TABLE IF NOT EXISTS holidays (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL,
		date TEXT NOT NULL,
		name TEXT NOT NULL,
		type TEXT DEFAULT 'national',
		location TEXT DEFAULT '',
		UNIQUE(year, date, type, location)
	);

	-- Chat history for AI interactions
	CREATE TABLE IF NOT EXISTS chat_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Insert default settings if not exist
	INSERT OR IGNORE INTO settings (key, value) VALUES 
		('openai_api_key', ''),
		('ai_provider', 'github'),
		('ai_model', 'openai/gpt-4o-mini'),
		('backend_port', '8080'),
		('frontend_port', '5173'),
		('default_work_week', '["monday","tuesday","wednesday","thursday","friday"]'),
		('default_vacation_days', '22'),
		('default_optimization_strategy', 'balanced'),
		('work_city', ''),
		('calendarific_api_key', '');
	`

	_, err := db.Exec(schema)
	if err != nil {
		return err
	}

	// Run migrations for existing databases
	migrations := []string{
		// Add reserved_days column if it doesn't exist
		`ALTER TABLE year_config ADD COLUMN reserved_days INTEGER DEFAULT 0;`,
		// Add optimizer_notes column if it doesn't exist
		`ALTER TABLE year_config ADD COLUMN optimizer_notes TEXT DEFAULT '';`,
		// Add location column to holidays if it doesn't exist
		`ALTER TABLE holidays ADD COLUMN location TEXT DEFAULT '';`,
	}

	for _, migration := range migrations {
		// Ignore errors (column may already exist)
		db.Exec(migration)
	}

	return nil
}
