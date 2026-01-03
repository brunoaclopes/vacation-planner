package holidays

import (
	"database/sql"
	"log"
	"sync"
	"time"
)

// HolidayStatus represents the current status of holiday data
type HolidayStatus struct {
	Year              int       `json:"year"`
	NationalLoaded    bool      `json:"national_loaded"`
	MunicipalLoaded   bool      `json:"municipal_loaded"`
	NationalError     string    `json:"national_error,omitempty"`
	MunicipalError    string    `json:"municipal_error,omitempty"`
	LastUpdated       time.Time `json:"last_updated"`
	RetryCount        int       `json:"retry_count"`
	MaxRetries        int       `json:"max_retries"`
	NextRetry         time.Time `json:"next_retry,omitempty"`
	IsRetrying        bool      `json:"is_retrying"`
}

// HolidayService manages holiday data with persistence and background retries
type HolidayService struct {
	db              *sql.DB
	status          map[int]*HolidayStatus
	statusMux       sync.RWMutex
	stopRetry       map[int]chan struct{}
	stopRetryMux    sync.Mutex
	maxRetries      int
	retryInterval   time.Duration
}

// NewHolidayService creates a new HolidayService
func NewHolidayService(db *sql.DB) *HolidayService {
	return &HolidayService{
		db:            db,
		status:        make(map[int]*HolidayStatus),
		stopRetry:     make(map[int]chan struct{}),
		maxRetries:    5,
		retryInterval: 30 * time.Second,
	}
}

// SetRetryConfig sets the retry configuration
func (s *HolidayService) SetRetryConfig(maxRetries int, interval time.Duration) {
	s.maxRetries = maxRetries
	s.retryInterval = interval
}

// GetStatus returns the current status for a year
func (s *HolidayService) GetStatus(year int) *HolidayStatus {
	s.statusMux.RLock()
	defer s.statusMux.RUnlock()
	
	if status, ok := s.status[year]; ok {
		return status
	}
	return nil
}

// GetAllStatuses returns status for all years with issues
func (s *HolidayService) GetAllStatuses() map[int]*HolidayStatus {
	s.statusMux.RLock()
	defer s.statusMux.RUnlock()
	
	result := make(map[int]*HolidayStatus)
	for year, status := range s.status {
		result[year] = status
	}
	return result
}

// LoadHolidaysForYear loads holidays from DB or fetches from API
func (s *HolidayService) LoadHolidaysForYear(year int, city string) ([]PortugueseHoliday, error) {
	// First, try to load from database
	dbHolidays, hasNational, hasMunicipal := s.loadFromDatabase(year, city)
	
	// Initialize status
	s.statusMux.Lock()
	if s.status[year] == nil {
		s.status[year] = &HolidayStatus{
			Year:       year,
			MaxRetries: s.maxRetries,
		}
	}
	status := s.status[year]
	s.statusMux.Unlock()
	
	// If we have national holidays in DB, use them
	if hasNational {
		status.NationalLoaded = true
		status.NationalError = ""
	}
	
	// If we have municipal holidays for this city in DB, use them
	if hasMunicipal {
		status.MunicipalLoaded = true
		status.MunicipalError = ""
	}
	
	// If we have data from DB, return it (we'll refresh in background if needed)
	if len(dbHolidays) > 0 {
		status.LastUpdated = time.Now()
		
		// Start background refresh if data might be stale (older than 24 hours)
		go s.refreshInBackground(year, city, !hasNational, !hasMunicipal && city != "")
		
		return dbHolidays, nil
	}
	
	// No data in DB, need to fetch from API
	return s.fetchAndSave(year, city)
}

// loadFromDatabase loads holidays from the database
func (s *HolidayService) loadFromDatabase(year int, city string) ([]PortugueseHoliday, bool, bool) {
	var holidays []PortugueseHoliday
	hasNational := false
	hasMunicipal := false
	
	query := `SELECT date, name, type, COALESCE(location, '') as location FROM holidays WHERE year = ?`
	rows, err := s.db.Query(query, year)
	if err != nil {
		log.Printf("Error loading holidays from DB: %v", err)
		return nil, false, false
	}
	defer rows.Close()
	
	for rows.Next() {
		var h PortugueseHoliday
		if err := rows.Scan(&h.Date, &h.Name, &h.Type, &h.Location); err != nil {
			continue
		}
		
		if h.Type == "national" {
			hasNational = true
			holidays = append(holidays, h)
		} else if h.Type == "municipal" {
			if city == "" || containsCity(h.Location, city) {
				hasMunicipal = true
				holidays = append(holidays, h)
			}
		}
	}
	
	return holidays, hasNational, hasMunicipal
}

// fetchAndSave fetches holidays from API and saves to database
func (s *HolidayService) fetchAndSave(year int, city string) ([]PortugueseHoliday, error) {
	var allHolidays []PortugueseHoliday
	
	s.statusMux.Lock()
	status := s.status[year]
	s.statusMux.Unlock()
	
	// Fetch national holidays
	nationalHolidays, err := fetchNationalHolidays(year)
	if err != nil {
		log.Printf("Warning: Failed to fetch national holidays: %v", err)
		status.NationalError = err.Error()
		status.NationalLoaded = false
		
		// Use fallback
		nationalHolidays = getFallbackNationalHolidays(year)
		
		// Start background retry
		s.startBackgroundRetry(year, city, true, false)
	} else {
		status.NationalLoaded = true
		status.NationalError = ""
		
		// Save to database
		s.saveHolidaysToDatabase(year, nationalHolidays)
	}
	allHolidays = append(allHolidays, nationalHolidays...)
	
	// Fetch municipal holidays if city is specified
	if city != "" {
		municipalHolidays, err := fetchMunicipalHolidays(year)
		if err != nil {
			log.Printf("Warning: Failed to fetch municipal holidays: %v", err)
			status.MunicipalError = err.Error()
			status.MunicipalLoaded = false
			
			// Start background retry for municipal
			s.startBackgroundRetry(year, city, false, true)
		} else {
			status.MunicipalLoaded = true
			status.MunicipalError = ""
			
			// Save municipal holidays to database
			s.saveHolidaysToDatabase(year, municipalHolidays)
			
			// Filter for the specific city
			for _, mh := range municipalHolidays {
				if containsCity(mh.Location, city) {
					allHolidays = append(allHolidays, mh)
				}
			}
		}
	}
	
	status.LastUpdated = time.Now()
	
	return allHolidays, nil
}

// saveHolidaysToDatabase saves holidays to the database
func (s *HolidayService) saveHolidaysToDatabase(year int, holidays []PortugueseHoliday) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO holidays (year, date, name, type, location) 
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	
	for _, h := range holidays {
		_, err := stmt.Exec(year, h.Date, h.Name, h.Type, h.Location)
		if err != nil {
			log.Printf("Error saving holiday to DB: %v", err)
		}
	}
	
	return tx.Commit()
}

// refreshInBackground refreshes holiday data in background
func (s *HolidayService) refreshInBackground(year int, city string, refreshNational, refreshMunicipal bool) {
	if !refreshNational && !refreshMunicipal {
		return
	}
	
	// Check if data needs refresh (check last_updated in status)
	s.statusMux.RLock()
	status := s.status[year]
	s.statusMux.RUnlock()
	
	if status == nil {
		return
	}
	
	// If last update was less than 1 hour ago, skip
	if time.Since(status.LastUpdated) < time.Hour {
		return
	}
	
	if refreshNational {
		nationalHolidays, err := fetchNationalHolidays(year)
		if err == nil {
			s.saveHolidaysToDatabase(year, nationalHolidays)
			s.statusMux.Lock()
			status.NationalLoaded = true
			status.NationalError = ""
			s.statusMux.Unlock()
			log.Printf("Background refresh: National holidays for %d updated", year)
		}
	}
	
	if refreshMunicipal && city != "" {
		municipalHolidays, err := fetchMunicipalHolidays(year)
		if err == nil {
			s.saveHolidaysToDatabase(year, municipalHolidays)
			s.statusMux.Lock()
			status.MunicipalLoaded = true
			status.MunicipalError = ""
			s.statusMux.Unlock()
			log.Printf("Background refresh: Municipal holidays for %d updated", year)
		}
	}
}

// startBackgroundRetry starts background retry for failed API calls
func (s *HolidayService) startBackgroundRetry(year int, city string, retryNational, retryMunicipal bool) {
	s.stopRetryMux.Lock()
	// Stop any existing retry goroutine for this year
	if stopChan, exists := s.stopRetry[year]; exists {
		close(stopChan)
	}
	stopChan := make(chan struct{})
	s.stopRetry[year] = stopChan
	s.stopRetryMux.Unlock()
	
	s.statusMux.Lock()
	status := s.status[year]
	status.RetryCount = 0
	status.IsRetrying = true
	status.NextRetry = time.Now().Add(s.retryInterval)
	s.statusMux.Unlock()
	
	go func() {
		ticker := time.NewTicker(s.retryInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-stopChan:
				s.statusMux.Lock()
				status.IsRetrying = false
				s.statusMux.Unlock()
				return
			case <-ticker.C:
				s.statusMux.Lock()
				status.RetryCount++
				currentRetry := status.RetryCount
				s.statusMux.Unlock()
				
				if currentRetry > s.maxRetries {
					log.Printf("Max retries reached for year %d, stopping background retry", year)
					s.statusMux.Lock()
					status.IsRetrying = false
					s.statusMux.Unlock()
					return
				}
				
				log.Printf("Background retry %d/%d for year %d holidays", currentRetry, s.maxRetries, year)
				
				allSuccess := true
				
				if retryNational && status.NationalError != "" {
					nationalHolidays, err := fetchNationalHolidays(year)
					if err != nil {
						log.Printf("Retry failed for national holidays: %v", err)
						allSuccess = false
						s.statusMux.Lock()
						status.NationalError = err.Error()
						status.NextRetry = time.Now().Add(s.retryInterval)
						s.statusMux.Unlock()
					} else {
						s.saveHolidaysToDatabase(year, nationalHolidays)
						s.statusMux.Lock()
						status.NationalLoaded = true
						status.NationalError = ""
						s.statusMux.Unlock()
						log.Printf("National holidays for %d loaded successfully on retry", year)
						retryNational = false
					}
				}
				
				if retryMunicipal && status.MunicipalError != "" {
					municipalHolidays, err := fetchMunicipalHolidays(year)
					if err != nil {
						log.Printf("Retry failed for municipal holidays: %v", err)
						allSuccess = false
						s.statusMux.Lock()
						status.MunicipalError = err.Error()
						status.NextRetry = time.Now().Add(s.retryInterval)
						s.statusMux.Unlock()
					} else {
						s.saveHolidaysToDatabase(year, municipalHolidays)
						s.statusMux.Lock()
						status.MunicipalLoaded = true
						status.MunicipalError = ""
						s.statusMux.Unlock()
						log.Printf("Municipal holidays for %d loaded successfully on retry", year)
						retryMunicipal = false
					}
				}
				
				// If all succeeded, stop retrying
				if allSuccess || (!retryNational && !retryMunicipal) {
					s.statusMux.Lock()
					status.IsRetrying = false
					status.LastUpdated = time.Now()
					s.statusMux.Unlock()
					return
				}
			}
		}
	}()
}

// StopAllRetries stops all background retry goroutines
func (s *HolidayService) StopAllRetries() {
	s.stopRetryMux.Lock()
	defer s.stopRetryMux.Unlock()
	
	for year, stopChan := range s.stopRetry {
		close(stopChan)
		delete(s.stopRetry, year)
	}
}

// ClearStatus clears the status for a year (useful when manually refreshing)
func (s *HolidayService) ClearStatus(year int) {
	s.statusMux.Lock()
	delete(s.status, year)
	s.statusMux.Unlock()
	
	s.stopRetryMux.Lock()
	if stopChan, exists := s.stopRetry[year]; exists {
		close(stopChan)
		delete(s.stopRetry, year)
	}
	s.stopRetryMux.Unlock()
}

// ForceRefresh forces a refresh of holidays for a year
func (s *HolidayService) ForceRefresh(year int, city string) ([]PortugueseHoliday, error) {
	// Clear existing status and stop any retries
	s.ClearStatus(year)
	
	// Delete from database
	_, err := s.db.Exec(`DELETE FROM holidays WHERE year = ?`, year)
	if err != nil {
		log.Printf("Error clearing holidays from DB: %v", err)
	}
	
	// Clear memory cache
	ClearCacheForYear(year)
	
	// Initialize new status
	s.statusMux.Lock()
	s.status[year] = &HolidayStatus{
		Year:       year,
		MaxRetries: s.maxRetries,
	}
	s.statusMux.Unlock()
	
	// Fetch fresh data
	return s.fetchAndSave(year, city)
}

// ToJSON returns the status as JSON for API responses
func (s *HolidayStatus) ToJSON() map[string]interface{} {
	result := map[string]interface{}{
		"year":             s.Year,
		"national_loaded":  s.NationalLoaded,
		"municipal_loaded": s.MunicipalLoaded,
		"last_updated":     s.LastUpdated.Format(time.RFC3339),
		"retry_count":      s.RetryCount,
		"max_retries":      s.MaxRetries,
		"is_retrying":      s.IsRetrying,
	}
	
	if s.NationalError != "" {
		result["national_error"] = s.NationalError
	}
	if s.MunicipalError != "" {
		result["municipal_error"] = s.MunicipalError
	}
	if s.IsRetrying && !s.NextRetry.IsZero() {
		result["next_retry"] = s.NextRetry.Format(time.RFC3339)
	}
	
	return result
}

// HasErrors returns true if there are any errors
func (s *HolidayStatus) HasErrors() bool {
	return s.NationalError != "" || s.MunicipalError != ""
}
