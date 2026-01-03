package holidays

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// PortugueseHoliday represents a Portuguese holiday
type PortugueseHoliday struct {
	Date     string `json:"date"`
	Name     string `json:"name"`
	Type     string `json:"type"`     // "national" or "municipal"
	Location string `json:"location"` // City/location for municipal holidays
}

// NagerHoliday represents a holiday from the Nager.Date API
type NagerHoliday struct {
	Date      string   `json:"date"`
	LocalName string   `json:"localName"`
	Name      string   `json:"name"`
	Fixed     bool     `json:"fixed"`
	Global    bool     `json:"global"`
	Counties  []string `json:"counties"`
	Types     []string `json:"types"`
}

// CalendarificResponse represents the response from Calendarific API
type CalendarificResponse struct {
	Meta struct {
		Code int `json:"code"`
	} `json:"meta"`
	Response struct {
		Holidays []CalendarificHoliday `json:"holidays"`
	} `json:"response"`
}

// CalendarificHoliday represents a holiday from Calendarific API
type CalendarificHoliday struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Date        struct {
		ISO string `json:"iso"`
	} `json:"date"`
	Type        []string    `json:"type"`
	PrimaryType string      `json:"primary_type"`
	Locations   string      `json:"locations"`
	States      interface{} `json:"states"` // Can be "All" string or array of state objects
}

var (
	// Cache for API responses
	holidayCache    = make(map[string][]PortugueseHoliday) // key: "year" or "year:city"
	holidayCacheMux sync.RWMutex

	// API configuration
	calendarificAPIKey string
	apiConfigMux       sync.RWMutex
)

const (
	nagerAPIURL       = "https://date.nager.at/api/v3/publicholidays/%d/PT"
	calendarificURL   = "https://calendarific.com/api/v2/holidays"
)

// SetCalendarificAPIKey sets the API key for Calendarific (for municipal holidays)
func SetCalendarificAPIKey(apiKey string) {
	apiConfigMux.Lock()
	defer apiConfigMux.Unlock()
	calendarificAPIKey = apiKey
}

// GetCalendarificAPIKey returns the current API key
func GetCalendarificAPIKey() string {
	apiConfigMux.RLock()
	defer apiConfigMux.RUnlock()
	return calendarificAPIKey
}

// fetchNationalHolidays fetches national holidays from the Nager.Date API
func fetchNationalHolidays(year int) ([]PortugueseHoliday, error) {
	url := fmt.Sprintf(nagerAPIURL, year)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch holidays from API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	var nagerHolidays []NagerHoliday
	if err := json.Unmarshal(body, &nagerHolidays); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	var holidays []PortugueseHoliday
	for _, nh := range nagerHolidays {
		// Only include global (national) holidays and public holidays
		isPublic := false
		for _, t := range nh.Types {
			if t == "Public" {
				isPublic = true
				break
			}
		}

		if nh.Global && isPublic {
			holidays = append(holidays, PortugueseHoliday{
				Date: nh.Date,
				Name: nh.LocalName,
				Type: "national",
			})
		}
	}

	return holidays, nil
}

// fetchMunicipalHolidays fetches municipal/local holidays from Calendarific API
func fetchMunicipalHolidays(year int) ([]PortugueseHoliday, error) {
	apiKey := GetCalendarificAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("calendarific API key not configured")
	}

	url := fmt.Sprintf("%s?api_key=%s&country=PT&year=%d&type=local", calendarificURL, apiKey, year)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch municipal holidays: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Calendarific API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	var calResponse CalendarificResponse
	if err := json.Unmarshal(body, &calResponse); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	var holidays []PortugueseHoliday
	for _, ch := range calResponse.Response.Holidays {
		// Check if it's a local/municipal holiday
		isLocal := false
		for _, t := range ch.Type {
			if t == "Local holiday" || t == "Local" || t == "Common local holiday" {
				isLocal = true
				break
			}
		}

		if isLocal {
			location := ch.Locations
			// Parse states if it's an array
			if location == "" || location == "All" {
				if statesArray, ok := ch.States.([]interface{}); ok && len(statesArray) > 0 {
					if stateMap, ok := statesArray[0].(map[string]interface{}); ok {
						if name, ok := stateMap["name"].(string); ok {
							location = name
						}
					}
				}
			}

			// Skip if no specific location
			if location == "" || location == "All" {
				continue
			}

			name := fmt.Sprintf("%s (%s)", ch.Name, location)

			holidays = append(holidays, PortugueseHoliday{
				Date:     ch.Date.ISO,
				Name:     name,
				Type:     "municipal",
				Location: location,
			})
		}
	}

	return holidays, nil
}

// getFallbackNationalHolidays returns calculated holidays as fallback when API fails
func getFallbackNationalHolidays(year int) []PortugueseHoliday {
	holidays := []PortugueseHoliday{
		{Date: formatDate(year, 1, 1), Name: "Ano Novo", Type: "national"},
		{Date: formatDate(year, 4, 25), Name: "Dia da Liberdade", Type: "national"},
		{Date: formatDate(year, 5, 1), Name: "Dia do Trabalhador", Type: "national"},
		{Date: formatDate(year, 6, 10), Name: "Dia de Portugal", Type: "national"},
		{Date: formatDate(year, 8, 15), Name: "Assunção de Nossa Senhora", Type: "national"},
		{Date: formatDate(year, 10, 5), Name: "Implantação da República", Type: "national"},
		{Date: formatDate(year, 11, 1), Name: "Dia de Todos os Santos", Type: "national"},
		{Date: formatDate(year, 12, 1), Name: "Restauração da Independência", Type: "national"},
		{Date: formatDate(year, 12, 8), Name: "Imaculada Conceição", Type: "national"},
		{Date: formatDate(year, 12, 25), Name: "Natal", Type: "national"},
	}

	// Calculate Easter-dependent holidays
	easter := calculateEaster(year)

	goodFriday := easter.AddDate(0, 0, -2)
	holidays = append(holidays, PortugueseHoliday{
		Date: goodFriday.Format("2006-01-02"),
		Name: "Sexta-feira Santa",
		Type: "national",
	})

	holidays = append(holidays, PortugueseHoliday{
		Date: easter.Format("2006-01-02"),
		Name: "Domingo de Páscoa",
		Type: "national",
	})

	corpusChristi := easter.AddDate(0, 0, 60)
	holidays = append(holidays, PortugueseHoliday{
		Date: corpusChristi.Format("2006-01-02"),
		Name: "Corpo de Deus",
		Type: "national",
	})

	return holidays
}

// GetPortugueseHolidays returns all Portuguese national holidays for a given year
func GetPortugueseHolidays(year int) []PortugueseHoliday {
	return GetPortugueseHolidaysWithCity(year, "")
}

// GetPortugueseHolidaysWithCity returns all Portuguese holidays including municipal ones for a city
func GetPortugueseHolidaysWithCity(year int, city string) []PortugueseHoliday {
	cacheKey := fmt.Sprintf("%d", year)
	if city != "" {
		cacheKey = fmt.Sprintf("%d:%s", year, city)
	}

	// Check cache first
	holidayCacheMux.RLock()
	cachedHolidays, found := holidayCache[cacheKey]
	holidayCacheMux.RUnlock()

	if found {
		return cachedHolidays
	}

	// Fetch national holidays
	nationalHolidays, err := fetchNationalHolidays(year)
	if err != nil {
		fmt.Printf("Warning: Failed to fetch holidays from API: %v. Using fallback.\n", err)
		nationalHolidays = getFallbackNationalHolidays(year)
	}

	// Create combined holidays list
	holidays := make([]PortugueseHoliday, len(nationalHolidays))
	copy(holidays, nationalHolidays)

	// Fetch municipal holidays if city is specified
	if city != "" {
		municipalHolidays, err := fetchMunicipalHolidays(year)
		if err != nil {
			fmt.Printf("Warning: Failed to fetch municipal holidays: %v\n", err)
		} else {
			// Filter for the specific city
			for _, mh := range municipalHolidays {
				// Check if this holiday is for the specified city
				if containsCity(mh.Location, city) {
					holidays = append(holidays, mh)
				}
			}
		}
	}

	// Cache the result
	holidayCacheMux.Lock()
	holidayCache[cacheKey] = holidays
	holidayCacheMux.Unlock()

	return holidays
}

// FetchAndCacheHolidays fetches holidays for a year and caches them
// Call this on app start or when year changes
func FetchAndCacheHolidays(year int) error {
	// Clear cache for this year
	holidayCacheMux.Lock()
	for key := range holidayCache {
		if key == fmt.Sprintf("%d", year) || len(key) > 4 && key[:4] == fmt.Sprintf("%d", year) {
			delete(holidayCache, key)
		}
	}
	holidayCacheMux.Unlock()

	// Fetch national holidays
	_, err := fetchNationalHolidays(year)
	if err != nil {
		return fmt.Errorf("failed to fetch national holidays: %w", err)
	}

	// Fetch all municipal holidays (they'll be filtered by city later)
	_, err = fetchMunicipalHolidays(year)
	if err != nil {
		// Not critical, just log
		fmt.Printf("Warning: Could not fetch municipal holidays: %v\n", err)
	}

	return nil
}

// GetAvailableCities returns cities that have municipal holidays
// This fetches from Calendarific API to get actual Portuguese municipalities
func GetAvailableCities() []string {
	// Portuguese municipalities with known municipal holidays
	// These are the main cities - the actual holidays come from the API
	cities := []string{
		"Lisboa",
		"Porto",
		"Braga",
		"Coimbra",
		"Setúbal",
		"Funchal",
		"Aveiro",
		"Viseu",
		"Leiria",
		"Faro",
		"Évora",
		"Guimarães",
		"Vila Nova de Gaia",
		"Matosinhos",
		"Almada",
		"Oeiras",
		"Cascais",
		"Sintra",
		"Loures",
		"Amadora",
		"Gondomar",
		"Maia",
		"Santarém",
		"Beja",
		"Castelo Branco",
		"Portalegre",
		"Vila Real",
		"Bragança",
		"Viana do Castelo",
		"Guarda",
		"Ponta Delgada",
		"Angra do Heroísmo",
		"Horta",
	}
	sort.Strings(cities)
	return cities
}

// ClearCache clears the holiday cache (useful for testing or forcing refresh)
func ClearCache() {
	holidayCacheMux.Lock()
	holidayCache = make(map[string][]PortugueseHoliday)
	holidayCacheMux.Unlock()
}

// ClearCacheForYear clears the holiday cache for a specific year
func ClearCacheForYear(year int) {
	holidayCacheMux.Lock()
	defer holidayCacheMux.Unlock()
	
	yearPrefix := fmt.Sprintf("%d", year)
	for key := range holidayCache {
		if key == yearPrefix || (len(key) > len(yearPrefix) && key[:len(yearPrefix)+1] == yearPrefix+":") {
			delete(holidayCache, key)
		}
	}
}

// normalizeCity normalizes city name for comparison
func normalizeCity(city string) string {
	return strings.ToLower(strings.TrimSpace(city))
}

// containsCity checks if a holiday location matches the city
func containsCity(holidayLocation, city string) bool {
	locationLower := strings.ToLower(holidayLocation)
	cityLower := strings.ToLower(city)
	
	// Direct match
	if locationLower == cityLower {
		return true
	}
	
	// Split location by comma to check each city separately
	// e.g., "Porto, Braga" -> ["Porto", "Braga"]
	locations := strings.Split(locationLower, ",")
	for _, loc := range locations {
		loc = strings.TrimSpace(loc)
		// Exact match on each location part
		if loc == cityLower {
			return true
		}
	}
	
	return false
}

// calculateEaster calculates Easter Sunday for a given year using the Anonymous Gregorian algorithm
func calculateEaster(year int) time.Time {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := ((h + l - 7*m + 114) % 31) + 1

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func formatDate(year, month, day int) string {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
}

// IsHoliday checks if a given date is a holiday
func IsHoliday(date time.Time, holidays []PortugueseHoliday) (bool, string) {
	dateStr := date.Format("2006-01-02")
	for _, h := range holidays {
		if h.Date == dateStr {
			return true, h.Name
		}
	}
	return false, ""
}
