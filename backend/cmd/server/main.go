package main

import (
	"log"
	"os"
	"time"

	"github.com/bruno.lopes/calendar/backend/internal/api"
	"github.com/bruno.lopes/calendar/backend/internal/database"
	"github.com/bruno.lopes/calendar/backend/internal/holidays"
)

func main() {
	// Initialize database
	db, err := database.Initialize("./data/calendar.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Load Calendarific API key from settings
	var calendarificKey string
	db.QueryRow(`SELECT value FROM settings WHERE key = 'calendarific_api_key'`).Scan(&calendarificKey)
	if calendarificKey != "" {
		holidays.SetCalendarificAPIKey(calendarificKey)
		log.Println("Calendarific API key loaded from settings")
	}

	// Create holiday service for startup pre-fetch
	holidayService := holidays.NewHolidayService(db)
	holidayService.SetRetryConfig(5, 30*time.Second) // 5 retries, 30 second interval

	// Get work city from settings
	var workCity string
	db.QueryRow(`SELECT value FROM settings WHERE key = 'work_city'`).Scan(&workCity)

	// Pre-fetch holidays for current year on startup (non-blocking)
	currentYear := time.Now().Year()
	log.Printf("Loading holidays for year %d...", currentYear)
	
	go func() {
		_, err := holidayService.LoadHolidaysForYear(currentYear, workCity)
		if err != nil {
			log.Printf("Warning: Failed to pre-fetch holidays: %v (will retry in background)", err)
		} else {
			log.Printf("Holidays for %d loaded successfully", currentYear)
		}
	}()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	server := api.NewServer(db)
	log.Printf("Starting server on port %s", port)
	if err := server.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
