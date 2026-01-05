package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"

	"github.com/bruno.lopes/calendar/backend/internal/holidays"
	"github.com/bruno.lopes/calendar/backend/internal/models"
	"github.com/bruno.lopes/calendar/backend/internal/optimizer"
)

type Handler struct {
	db             *sql.DB
	holidayService *holidays.HolidayService
}

// isHoliday checks if a given date string is a holiday
func (h *Handler) isHoliday(dateStr string, year int) bool {
	workCity := h.getWorkCity()
	holidayList := holidays.GetPortugueseHolidaysWithCity(year, workCity)
	for _, holiday := range holidayList {
		if holiday.Date == dateStr {
			return true
		}
	}
	return false
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		db:             db,
		holidayService: holidays.NewHolidayService(db),
	}
}

// getWorkCity returns the configured work city for municipal holidays
func (h *Handler) getWorkCity() string {
	var city string
	h.db.QueryRow(`SELECT value FROM settings WHERE key = 'work_city'`).Scan(&city)
	return city
}

// GetCalendar returns the full calendar for a year
func (h *Handler) GetCalendar(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	// Get or create year config
	config, err := h.getOrCreateYearConfig(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get holidays with work city for municipal holidays
	workCity := h.getWorkCity()
	holidayList := holidays.GetPortugueseHolidaysWithCity(year, workCity)
	
	// Store holidays in database
	for _, hol := range holidayList {
		h.db.Exec(`INSERT OR IGNORE INTO holidays (year, date, name, type) VALUES (?, ?, ?, ?)`,
			year, hol.Date, hol.Name, hol.Type)
	}

	// Get manual vacations
	manualVacations, _ := h.getVacations(year)

	// Get optimal vacations
	optimalVacations, _ := h.getOptimalVacations(year)

	// Build calendar days
	days := h.buildCalendarDays(year, config, holidayList, manualVacations, optimalVacations)

	// Calculate summary
	summary := h.calculateSummary(config.VacationDays, manualVacations, optimalVacations, holidayList)

	// Convert holidays to model
	var modelHolidays []models.Holiday
	for _, hol := range holidayList {
		modelHolidays = append(modelHolidays, models.Holiday{
			Year: year,
			Date: hol.Date,
			Name: hol.Name,
			Type: hol.Type,
		})
	}

	response := models.CalendarResponse{
		Year:             year,
		Config:           config,
		Days:             days,
		Holidays:         modelHolidays,
		ManualVacations:  manualVacations,
		OptimalVacations: optimalVacations,
		Summary:          summary,
	}

	c.JSON(http.StatusOK, response)
}

// OptimizeVacations calculates optimal vacation days
func (h *Handler) OptimizeVacations(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	config, err := h.getOrCreateYearConfig(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get manual vacations to exclude
	manualVacations, _ := h.getVacations(year)
	var manualDates []string
	for _, v := range manualVacations {
		manualDates = append(manualDates, v.Date)
	}

	// Calculate available days for optimizer (total - reserved - manual)
	availableDays := config.VacationDays - config.ReservedDays - len(manualDates)
	if availableDays < 0 {
		availableDays = 0
	}

	var blocks []models.VacationBlock

	// Check if using smart AI strategy
	if config.OptimizationStrategy == models.StrategySmart {
		blocks, err = h.smartOptimize(year, availableDays, config.WorkWeek, manualDates)
		if err != nil {
			// Fallback to balanced strategy if AI fails
			workCity := h.getWorkCity()
			opt := optimizer.NewOptimizerWithCity(year, availableDays, config.WorkWeek, models.StrategyBalanced, workCity)
			opt.SetManualVacations(manualDates)
			blocks = opt.Optimize()
		}
	} else {
		// Run regular optimizer with city-specific holidays
		workCity := h.getWorkCity()
		opt := optimizer.NewOptimizerWithCity(year, availableDays, config.WorkWeek, config.OptimizationStrategy, workCity)
		opt.SetManualVacations(manualDates)
		blocks = opt.Optimize()
	}

	// Clear previous optimal vacations
	h.db.Exec("DELETE FROM optimal_vacations WHERE year = ?", year)

	// Store new optimal vacations
	blockID := 1
	for _, block := range blocks {
		for _, date := range block.Dates {
			// Only store dates that require vacation days
			if !contains(block.Weekends, date) && !contains(block.Holidays, date) && !contains(manualDates, date) {
				h.db.Exec(`INSERT OR REPLACE INTO optimal_vacations (year, date, block_id, consecutive_days) VALUES (?, ?, ?, ?)`,
					year, date, blockID, block.TotalDays)
			}
		}
		blockID++
	}

	c.JSON(http.StatusOK, gin.H{
		"blocks": blocks,
		"message": "Optimization complete",
	})
}

// smartOptimize uses AI to find optimal vacation combinations
func (h *Handler) smartOptimize(year, availableDays int, workWeek, manualDates []string) ([]models.VacationBlock, error) {
	// Get API key and provider
	var apiKey string
	err := h.db.QueryRow("SELECT value FROM settings WHERE key = 'openai_api_key'").Scan(&apiKey)
	if err != nil || apiKey == "" {
		return nil, fmt.Errorf("API key not configured")
	}

	var aiProvider string
	h.db.QueryRow("SELECT value FROM settings WHERE key = 'ai_provider'").Scan(&aiProvider)
	if aiProvider == "" {
		aiProvider = "github"
	}

	var selectedModel string
	h.db.QueryRow("SELECT value FROM settings WHERE key = 'ai_model'").Scan(&selectedModel)
	if selectedModel == "" {
		selectedModel = "openai/gpt-4o-mini"
	}

	if aiProvider == "github" && !strings.Contains(selectedModel, "/") {
		selectedModel = "openai/" + selectedModel
	}

	// Get holidays
	workCity := h.getWorkCity()
	holidayList := holidays.GetPortugueseHolidaysWithCity(year, workCity)

	// Build context for AI
	var holidayInfo strings.Builder
	for _, h := range holidayList {
		date, _ := time.Parse("2006-01-02", h.Date)
		holidayInfo.WriteString(fmt.Sprintf("- %s (%s): %s\n", h.Date, date.Weekday().String(), h.Name))
	}

	var manualInfo string
	if len(manualDates) > 0 {
		manualInfo = fmt.Sprintf("Already scheduled vacation days (do NOT include these): %s\n", strings.Join(manualDates, ", "))
	}

	// Get optimizer notes from year config
	var optimizerNotes string
	h.db.QueryRow("SELECT COALESCE(optimizer_notes, '') FROM year_config WHERE year = ?", year).Scan(&optimizerNotes)
	var userNotesInfo string
	if optimizerNotes != "" {
		userNotesInfo = fmt.Sprintf("\nUSER PREFERENCES/NOTES (IMPORTANT - follow these instructions):\n%s\n", optimizerNotes)
	}

	// Determine weekend days (days not in work week)
	workDaySet := make(map[string]bool)
	for _, d := range workWeek {
		workDaySet[strings.ToLower(d)] = true
	}
	var weekendDays []string
	allDays := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	for _, d := range allDays {
		if !workDaySet[d] {
			weekendDays = append(weekendDays, d)
		}
	}

	prompt := fmt.Sprintf(`You are a vacation optimization expert. Find the BEST vacation days for year %d.

CONSTRAINTS:
- You have exactly %d vacation days to allocate
- Work days: %v (ONLY select dates that fall on these days!)
- Weekend/Off days: %v (these are FREE days off - NEVER select these as vacation days!)
- You must use ALL %d days - no more, no less
- CRITICAL: Only select dates that are WORK DAYS (not weekends, not holidays)
%s%s
HOLIDAYS (already days off, don't count as vacation days):
%s
CRITICAL OPTIMIZATION STRATEGY:
The goal is to MAXIMIZE total consecutive days off while using the MINIMUM vacation days.

KEY INSIGHT: Weekends (%v) are FREE! Combine them with holidays and vacation days to create long breaks.

EXAMPLES OF SMART OPTIMIZATION:
- If a holiday falls on Thursday, take Friday as vacation → get Thu+Fri+Sat+Sun = 4 days off for 1 vacation day!
- If a holiday falls on Tuesday, take Monday as vacation → get Sat+Sun+Mon+Tue = 4 days off for 1 vacation day!
- If holidays are on Thu and following Mon, take Fri as vacation → get Thu+Fri+Sat+Sun+Mon = 5 days off for 1 vacation day!

OPTIMIZATION PRIORITIES (in order):
1. BRIDGE holidays with adjacent weekends - this gives the best efficiency
2. Look for holidays on Tuesday/Thursday - taking Mon or Fri creates 4-day weekends with just 1 vacation day
3. Look for holidays near weekends that can be extended
4. Group vacation days to create week-long breaks when possible
5. Distribute throughout the year for work-life balance

IMPORTANT VALIDATION RULES:
- Each date MUST fall on a work day (%v)
- Each date must NOT be a weekend day (%v)
- Each date must NOT be a holiday

RESPOND WITH ONLY a JSON array of vacation day dates in YYYY-MM-DD format.
Example: ["2026-01-02", "2026-04-06", "2026-12-28"]

Analyze each holiday's day of the week and find the optimal bridging strategy.
Return EXACTLY %d dates as a JSON array, nothing else.`, year, availableDays, workWeek, weekendDays, availableDays, manualInfo, userNotesInfo, holidayInfo.String(), weekendDays, workWeek, weekendDays, availableDays)

	// Create AI client
	var client *openai.Client
	switch aiProvider {
	case "github":
		config := openai.DefaultConfig(apiKey)
		config.BaseURL = "https://models.github.ai/inference"
		client = openai.NewClientWithConfig(config)
	case "openai":
		client = openai.NewClient(apiKey)
	default:
		config := openai.DefaultConfig(apiKey)
		config.BaseURL = "https://models.github.ai/inference"
		client = openai.NewClientWithConfig(config)
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: selectedModel,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			Temperature: 0.3, // Lower temperature for more deterministic results
		},
	)

	if err != nil {
		return nil, fmt.Errorf("AI request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	// Parse AI response
	responseText := resp.Choices[0].Message.Content
	
	// Extract JSON array from response
	jsonRegex := regexp.MustCompile(`\[[\s\S]*?\]`)
	jsonMatch := jsonRegex.FindString(responseText)
	if jsonMatch == "" {
		return nil, fmt.Errorf("could not parse AI response")
	}

	var vacationDates []string
	if err := json.Unmarshal([]byte(jsonMatch), &vacationDates); err != nil {
		return nil, fmt.Errorf("failed to parse vacation dates: %w", err)
	}

	// Create work day lookup for validation
	workDayMap := make(map[string]bool)
	for _, d := range workWeek {
		workDayMap[strings.ToLower(d)] = true
	}

	// Create holiday lookup for validation
	holidayMap := make(map[string]bool)
	for _, hol := range holidayList {
		holidayMap[hol.Date] = true
	}

	// Filter out any invalid dates (weekends or holidays)
	var validDates []string
	for _, dateStr := range vacationDates {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		dayName := strings.ToLower(date.Weekday().String())
		// Skip if it's a weekend (not a work day)
		if !workDayMap[dayName] {
			continue
		}
		// Skip if it's a holiday
		if holidayMap[dateStr] {
			continue
		}
		validDates = append(validDates, dateStr)
	}

	// Convert dates to vacation blocks
	return h.datesToBlocks(year, validDates, holidayList, workWeek)
}

// datesToBlocks converts a list of vacation dates to VacationBlock structures
func (h *Handler) datesToBlocks(year int, vacationDates []string, holidayList []holidays.PortugueseHoliday, workWeek []string) ([]models.VacationBlock, error) {
	if len(vacationDates) == 0 {
		return nil, nil
	}

	// Create holiday lookup
	holidayMap := make(map[string]bool)
	for _, hol := range holidayList {
		holidayMap[hol.Date] = true
	}

	// Create work day lookup
	workDayMap := make(map[string]bool)
	for _, d := range workWeek {
		workDayMap[strings.ToLower(d)] = true
	}

	isWeekend := func(date time.Time) bool {
		dayName := strings.ToLower(date.Weekday().String())
		return !workDayMap[dayName]
	}

	// Sort vacation dates
	sort.Strings(vacationDates)

	// Group into consecutive blocks (including weekends and holidays)
	var blocks []models.VacationBlock
	
	for _, vacDateStr := range vacationDates {
		vacDate, err := time.Parse("2006-01-02", vacDateStr)
		if err != nil {
			continue
		}

		// Find or create block that this date extends
		added := false
		for i := range blocks {
			blockEnd, _ := time.Parse("2006-01-02", blocks[i].EndDate)
			
			// Check if this date extends the block (allowing for weekends/holidays in between)
			dayAfterBlock := blockEnd.AddDate(0, 0, 1)
			for !dayAfterBlock.After(vacDate) {
				dateStr := dayAfterBlock.Format("2006-01-02")
				if dayAfterBlock.Equal(vacDate) {
					// Extend the block
					blocks[i].EndDate = vacDateStr
					blocks[i].Dates = append(blocks[i].Dates, vacDateStr)
					blocks[i].TotalDays++
					blocks[i].VacationDaysUsed++
					added = true
					break
				} else if isWeekend(dayAfterBlock) || holidayMap[dateStr] {
					// Weekend or holiday - add to block and continue checking
					blocks[i].EndDate = dateStr
					blocks[i].Dates = append(blocks[i].Dates, dateStr)
					blocks[i].TotalDays++
					if isWeekend(dayAfterBlock) {
						blocks[i].Weekends = append(blocks[i].Weekends, dateStr)
					} else {
						blocks[i].Holidays = append(blocks[i].Holidays, dateStr)
					}
					dayAfterBlock = dayAfterBlock.AddDate(0, 0, 1)
				} else {
					// Gap - not part of this block
					break
				}
			}
			if added {
				break
			}
		}

		if !added {
			// Start new block, expanding backwards to include preceding weekends/holidays
			startDate := vacDate
			var preDates []string
			var preWeekends []string
			var preHolidays []string
			
			checkDate := vacDate.AddDate(0, 0, -1)
			for {
				dateStr := checkDate.Format("2006-01-02")
				if isWeekend(checkDate) {
					preDates = append([]string{dateStr}, preDates...)
					preWeekends = append([]string{dateStr}, preWeekends...)
					startDate = checkDate
					checkDate = checkDate.AddDate(0, 0, -1)
				} else if holidayMap[dateStr] {
					preDates = append([]string{dateStr}, preDates...)
					preHolidays = append([]string{dateStr}, preHolidays...)
					startDate = checkDate
					checkDate = checkDate.AddDate(0, 0, -1)
				} else {
					break
				}
			}

			block := models.VacationBlock{
				StartDate:        startDate.Format("2006-01-02"),
				EndDate:          vacDateStr,
				TotalDays:        len(preDates) + 1,
				VacationDaysUsed: 1,
				Dates:            append(preDates, vacDateStr),
				Weekends:         preWeekends,
				Holidays:         preHolidays,
			}
			blocks = append(blocks, block)
		}
	}

	// Expand blocks forward to include trailing weekends/holidays
	for i := range blocks {
		endDate, _ := time.Parse("2006-01-02", blocks[i].EndDate)
		checkDate := endDate.AddDate(0, 0, 1)
		
		for {
			dateStr := checkDate.Format("2006-01-02")
			if isWeekend(checkDate) {
				blocks[i].EndDate = dateStr
				blocks[i].Dates = append(blocks[i].Dates, dateStr)
				blocks[i].TotalDays++
				blocks[i].Weekends = append(blocks[i].Weekends, dateStr)
				checkDate = checkDate.AddDate(0, 0, 1)
			} else if holidayMap[dateStr] {
				blocks[i].EndDate = dateStr
				blocks[i].Dates = append(blocks[i].Dates, dateStr)
				blocks[i].TotalDays++
				blocks[i].Holidays = append(blocks[i].Holidays, dateStr)
				checkDate = checkDate.AddDate(0, 0, 1)
			} else {
				break
			}
		}
	}

	return blocks, nil
}

// GetVacations returns manual vacation days for a year
func (h *Handler) GetVacations(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	vacations, err := h.getVacations(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, vacations)
}

// AddVacation adds a manual vacation day
func (h *Handler) AddVacation(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	var input struct {
		Date string `json:"date" binding:"required"`
		Note string `json:"note"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the date is a holiday - can't set vacation on a holiday
	if h.isHoliday(input.Date, year) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot set vacation on a holiday"})
		return
	}

	_, err = h.db.Exec(`INSERT OR REPLACE INTO vacation_days (year, date, is_manual, note) VALUES (?, ?, TRUE, ?)`,
		year, input.Date, input.Note)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vacation day added"})
}

// RemoveVacation removes a vacation day
func (h *Handler) RemoveVacation(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	date := c.Param("date")

	_, err = h.db.Exec(`DELETE FROM vacation_days WHERE year = ? AND date = ?`, year, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vacation day removed"})
}

// ClearOptimizedVacations clears all optimized vacation days for a year
func (h *Handler) ClearOptimizedVacations(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	_, err = h.db.Exec(`DELETE FROM optimal_vacations WHERE year = ?`, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Optimized vacation days cleared"})
}

// GetVacationSuggestions uses AI to analyze manual vacation days and suggest improvements
func (h *Handler) GetVacationSuggestions(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	// Get language parameter (default to English)
	language := c.Query("language")
	if language == "" {
		language = "en"
	}

	// Get AI configuration
	var apiKey string
	err = h.db.QueryRow("SELECT value FROM settings WHERE key = 'openai_api_key'").Scan(&apiKey)
	if err != nil || apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key not configured"})
		return
	}

	var aiProvider string
	h.db.QueryRow("SELECT value FROM settings WHERE key = 'ai_provider'").Scan(&aiProvider)
	if aiProvider == "" {
		aiProvider = "github"
	}

	var selectedModel string
	h.db.QueryRow("SELECT value FROM settings WHERE key = 'ai_model'").Scan(&selectedModel)
	if selectedModel == "" {
		selectedModel = "openai/gpt-4o-mini"
	}

	if aiProvider == "github" && !strings.Contains(selectedModel, "/") {
		selectedModel = "openai/" + selectedModel
	}

	// Get year config
	config, _ := h.getOrCreateYearConfig(year)

	// Get manual vacations
	manualVacations, _ := h.getVacations(year)
	if len(manualVacations) == 0 {
		noVacationMsg := "You haven't set any manual vacation days yet. Add some vacation days first, then I can suggest improvements!"
		if language == "pt-PT" {
			noVacationMsg = "Ainda não definiu dias de férias manuais. Adicione alguns dias de férias primeiro, depois posso sugerir melhorias!"
		}
		c.JSON(http.StatusOK, gin.H{
			"suggestion": noVacationMsg,
		})
		return
	}

	// Get holidays
	workCity := h.getWorkCity()
	holidayList := holidays.GetPortugueseHolidaysWithCity(year, workCity)

	// Build holiday set for quick lookup
	holidaySet := make(map[string]bool)
	for _, hol := range holidayList {
		holidaySet[hol.Date] = true
	}

	// Build vacation set for quick lookup
	vacationSet := make(map[string]bool)
	for _, v := range manualVacations {
		vacationSet[v.Date] = true
	}

	// Build context
	var holidayInfo strings.Builder
	for _, hol := range holidayList {
		date, _ := time.Parse("2006-01-02", hol.Date)
		holidayInfo.WriteString(fmt.Sprintf("- %s (%s): %s\n", hol.Date, date.Weekday().String(), hol.Name))
	}

	// Get current date first
	today := time.Now()
	todayStr := today.Format("2006-01-02")

	var manualInfo strings.Builder
	manualInfo.WriteString(fmt.Sprintf("(Today is %s - only FUTURE dates can be moved)\n", todayStr))
	for _, v := range manualVacations {
		date, _ := time.Parse("2006-01-02", v.Date)
		if date.After(today) || date.Format("2006-01-02") == todayStr {
			manualInfo.WriteString(fmt.Sprintf("- %s (%s) - CAN BE MOVED\n", v.Date, date.Weekday().String()))
		} else {
			manualInfo.WriteString(fmt.Sprintf("- %s (%s) - IN THE PAST, cannot move\n", v.Date, date.Weekday().String()))
		}
	}

	// Determine weekend days
	workDaySet := make(map[string]bool)
	for _, d := range config.WorkWeek {
		workDaySet[strings.ToLower(d)] = true
	}
	var weekendDays []string
	allDays := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	for _, d := range allDays {
		if !workDaySet[d] {
			weekendDays = append(weekendDays, d)
		}
	}

	// Build list of bridge opportunity dates (work days adjacent to holidays/weekends)
	// These are the ONLY valid dates the AI should suggest
	
	// Helper function to calculate consecutive days off if we add a vacation on a specific date
	// IMPORTANT: Only counts weekends and holidays, NOT existing vacations (since we're moving them)
	calcBreak := func(vacDate time.Time) (int, string) {
		days := []string{}
		
		// Go backwards to find start of break
		for d := vacDate.AddDate(0, 0, -1); ; d = d.AddDate(0, 0, -1) {
			dStr := d.Format("2006-01-02")
			wdStr := strings.ToLower(d.Weekday().String())
			// Only count weekends and holidays, NOT existing vacations
			isOff := !workDaySet[wdStr] || holidaySet[dStr]
			if !isOff {
				break
			}
			days = append([]string{fmt.Sprintf("%s (%s)", dStr, d.Weekday().String()[:3])}, days...)
		}
		
		// Add the vacation day itself
		days = append(days, fmt.Sprintf("%s (%s, NEW)", vacDate.Format("2006-01-02"), vacDate.Weekday().String()[:3]))
		
		// Go forward to find end of break
		for d := vacDate.AddDate(0, 0, 1); ; d = d.AddDate(0, 0, 1) {
			dStr := d.Format("2006-01-02")
			wdStr := strings.ToLower(d.Weekday().String())
			// Only count weekends and holidays, NOT existing vacations
			isOff := !workDaySet[wdStr] || holidaySet[dStr]
			if !isOff {
				break
			}
			days = append(days, fmt.Sprintf("%s (%s)", dStr, d.Weekday().String()[:3]))
		}
		
		return len(days), strings.Join(days, " → ")
	}
	
	// Build bridge opportunities with pre-calculated break lengths
	type bridgeOpp struct {
		date      string
		weekday   string
		holiday   string
		breakDays int
		breakList string
	}
	var opportunities []bridgeOpp
	
	for _, hol := range holidayList {
		holDate, _ := time.Parse("2006-01-02", hol.Date)
		if holDate.Before(today) {
			continue
		}
		
		for offset := -3; offset <= 3; offset++ {
			if offset == 0 {
				continue
			}
			checkDate := holDate.AddDate(0, 0, offset)
			checkDateStr := checkDate.Format("2006-01-02")
			weekdayStr := strings.ToLower(checkDate.Weekday().String())
			
			if workDaySet[weekdayStr] && !holidaySet[checkDateStr] && !vacationSet[checkDateStr] && checkDate.After(today) {
				breakDays, breakList := calcBreak(checkDate)
				if breakDays >= 3 { // Only include if it creates at least 3 days off
					opportunities = append(opportunities, bridgeOpp{
						date:      checkDateStr,
						weekday:   checkDate.Weekday().String(),
						holiday:   hol.Name,
						breakDays: breakDays,
						breakList: breakList,
					})
				}
			}
		}
	}
	
	// Sort by break days (descending) and deduplicate
	seen := make(map[string]bool)
	var bridgeOpportunities strings.Builder
	bridgeOpportunities.WriteString("PRE-CALCULATED BRIDGE OPPORTUNITIES (take 1 vacation day, get X days off):\n")
	for _, opp := range opportunities {
		if seen[opp.date] {
			continue
		}
		seen[opp.date] = true
		bridgeOpportunities.WriteString(fmt.Sprintf("- Take %s (%s) off → %d consecutive days: %s\n", 
			opp.date, opp.weekday, opp.breakDays, opp.breakList))
	}

	// Determine response language
	languageInstruction := "Respond in English."
	if language == "pt-PT" {
		languageInstruction = "Respond in Portuguese (Portugal). Use European Portuguese, not Brazilian Portuguese."
	}

	// Get current date for context
	todayWeekday := today.Weekday().String()

	prompt := fmt.Sprintf(`You are a vacation planning advisor.

%s

TODAY'S DATE: %s (%s) - ONLY suggest moving vacation days that are AFTER this date!

USER'S CURRENT VACATION DAYS:
%s

HOLIDAYS (already days off):
%s

%s

TASK: Look at the user's current vacation days. If any are "isolated" (not creating a long break), suggest moving them to one of the pre-calculated bridge opportunities above.

The bridge opportunities already show EXACTLY how many days off you get and which days are included. Just use those numbers directly - do not recalculate.

Format your response as:
- Brief assessment of current vacation placement (1-2 sentences)
- 2-3 suggestions: "Move [current vacation date] to [bridge date] to get [X] days off: [copy the day sequence from above]"

Keep it concise.`, languageInstruction, todayStr, todayWeekday, manualInfo.String(), holidayInfo.String(), bridgeOpportunities.String())

	// Create AI client
	var client *openai.Client
	switch aiProvider {
	case "github":
		aiConfig := openai.DefaultConfig(apiKey)
		aiConfig.BaseURL = "https://models.github.ai/inference"
		client = openai.NewClientWithConfig(aiConfig)
	case "openai":
		client = openai.NewClient(apiKey)
	default:
		aiConfig := openai.DefaultConfig(apiKey)
		aiConfig.BaseURL = "https://models.github.ai/inference"
		client = openai.NewClientWithConfig(aiConfig)
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: selectedModel,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			Temperature: 0.3,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI request failed: " + err.Error()})
		return
	}

	if len(resp.Choices) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No response from AI"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"suggestion": resp.Choices[0].Message.Content,
	})
}

// BulkUpdateVacations updates multiple vacation days at once
func (h *Handler) BulkUpdateVacations(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	var input struct {
		Add    []string `json:"add"`
		Remove []string `json:"remove"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Remove vacations
	for _, date := range input.Remove {
		h.db.Exec(`DELETE FROM vacation_days WHERE year = ? AND date = ?`, year, date)
	}

	// Add vacations
	for _, date := range input.Add {
		h.db.Exec(`INSERT OR REPLACE INTO vacation_days (year, date, is_manual) VALUES (?, ?, TRUE)`, year, date)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vacations updated"})
}

// GetHolidays returns holidays for a year
func (h *Handler) GetHolidays(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	workCity := h.getWorkCity()
	
	// Use the holiday service which handles DB persistence and retries
	holidayList, err := h.holidayService.LoadHolidaysForYear(year, workCity)
	if err != nil {
		// Even on error, we should have fallback data
		holidayList = holidays.GetPortugueseHolidaysWithCity(year, workCity)
	}
	
	c.JSON(http.StatusOK, holidayList)
}

// GetHolidayStatus returns the current status of holiday data loading
func (h *Handler) GetHolidayStatus(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}
	
	status := h.holidayService.GetStatus(year)
	if status == nil {
		c.JSON(http.StatusOK, gin.H{
			"year":             year,
			"national_loaded":  true,
			"municipal_loaded": true,
			"has_errors":       false,
		})
		return
	}
	
	response := status.ToJSON()
	response["has_errors"] = status.HasErrors()
	c.JSON(http.StatusOK, response)
}

// GetAllHolidayStatuses returns status for all years
func (h *Handler) GetAllHolidayStatuses(c *gin.Context) {
	statuses := h.holidayService.GetAllStatuses()
	
	result := make([]map[string]interface{}, 0)
	for _, status := range statuses {
		if status.HasErrors() {
			statusJSON := status.ToJSON()
			statusJSON["has_errors"] = true
			result = append(result, statusJSON)
		}
	}
	
	c.JSON(http.StatusOK, result)
}

// GetAvailableCities returns all available cities for municipal holidays
func (h *Handler) GetAvailableCities(c *gin.Context) {
	cities := holidays.GetAvailableCities()
	c.JSON(http.StatusOK, cities)
}

// GetYearConfig returns configuration for a year
func (h *Handler) GetYearConfig(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	config, err := h.getOrCreateYearConfig(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateYearConfig updates configuration for a year
func (h *Handler) UpdateYearConfig(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	var input struct {
		VacationDays         *int     `json:"vacation_days"`
		ReservedDays         *int     `json:"reserved_days"`
		OptimizationStrategy *string  `json:"optimization_strategy"`
		WorkWeek             []string `json:"work_week"`
		OptimizerNotes       *string  `json:"optimizer_notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current config
	config, _ := h.getOrCreateYearConfig(year)

	// Update fields if provided
	if input.VacationDays != nil {
		config.VacationDays = *input.VacationDays
	}
	if input.ReservedDays != nil {
		config.ReservedDays = *input.ReservedDays
	}
	if input.OptimizationStrategy != nil {
		config.OptimizationStrategy = *input.OptimizationStrategy
	}
	if len(input.WorkWeek) > 0 {
		config.WorkWeek = input.WorkWeek
	}
	if input.OptimizerNotes != nil {
		config.OptimizerNotes = *input.OptimizerNotes
	}

	workWeekJSON, _ := json.Marshal(config.WorkWeek)

	_, err = h.db.Exec(`UPDATE year_config SET vacation_days = ?, reserved_days = ?, optimization_strategy = ?, work_week = ?, optimizer_notes = ?, updated_at = CURRENT_TIMESTAMP WHERE year = ?`,
		config.VacationDays, config.ReservedDays, config.OptimizationStrategy, string(workWeekJSON), config.OptimizerNotes, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// CopyYearConfig copies configuration from one year to another
func (h *Handler) CopyYearConfig(c *gin.Context) {
	yearStr := c.Param("year")
	sourceYearStr := c.Param("sourceYear")

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	sourceYear, err := strconv.Atoi(sourceYearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid source year"})
		return
	}

	sourceConfig, err := h.getOrCreateYearConfig(sourceYear)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	workWeekJSON, _ := json.Marshal(sourceConfig.WorkWeek)

	_, err = h.db.Exec(`INSERT OR REPLACE INTO year_config (year, vacation_days, optimization_strategy, work_week) VALUES (?, ?, ?, ?)`,
		year, sourceConfig.VacationDays, sourceConfig.OptimizationStrategy, string(workWeekJSON))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuration copied"})
}

// GetSettings returns all settings
func (h *Handler) GetSettings(c *gin.Context) {
	rows, err := h.db.Query("SELECT key, value FROM settings")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		rows.Scan(&key, &value)
		settings[key] = value
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateSettings updates multiple settings
func (h *Handler) UpdateSettings(c *gin.Context) {
	var input map[string]string
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for key, value := range input {
		h.db.Exec(`INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)`, key, value)
		
		// Update Calendarific API key if changed
		if key == "calendarific_api_key" {
			holidays.SetCalendarificAPIKey(value)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated"})
}

// GetSetting returns a single setting
func (h *Handler) GetSetting(c *gin.Context) {
	key := c.Param("key")

	var value string
	err := h.db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{key: value})
}

// UpdateSetting updates a single setting
func (h *Handler) UpdateSetting(c *gin.Context) {
	key := c.Param("key")

	var input struct {
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.db.Exec(`INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)`, key, input.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update Calendarific API key if changed
	if key == "calendarific_api_key" {
		holidays.SetCalendarificAPIKey(input.Value)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Setting updated"})
}

// RefreshHolidays clears cache and re-fetches holidays for a year
func (h *Handler) RefreshHolidays(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	workCity := h.getWorkCity()
	
	// Force refresh using the service (clears DB and memory cache)
	holidayList, err := h.holidayService.ForceRefresh(year, workCity)
	if err != nil {
		// Return whatever we have
		holidayList = holidays.GetPortugueseHolidaysWithCity(year, workCity)
	}
	
	status := h.holidayService.GetStatus(year)
	
	response := gin.H{
		"message":  "Holidays refreshed",
		"holidays": holidayList,
	}
	
	if status != nil && status.HasErrors() {
		response["status"] = status.ToJSON()
		response["has_errors"] = true
	}

	c.JSON(http.StatusOK, response)
}

// GetWorkWeekPresets returns available work week presets
func (h *Handler) GetWorkWeekPresets(c *gin.Context) {
	c.JSON(http.StatusOK, models.WorkWeekPresets)
}

// GetOptimizationStrategies returns available optimization strategies
func (h *Handler) GetOptimizationStrategies(c *gin.Context) {
	strategies := []map[string]string{
		{"id": models.StrategyBridgeHolidays, "name": "Bridge Holidays", "description": "Focus on creating bridges between holidays and weekends for efficient use of vacation days"},
		{"id": models.StrategyLongestBlocks, "name": "Longest Blocks", "description": "Focus on creating the longest possible consecutive vacation periods"},
		{"id": models.StrategyBalanced, "name": "Balanced", "description": "Balance between efficiency and length of vacation blocks"},
		{"id": models.StrategySmart, "name": "Smart (AI)", "description": "Use AI to find the optimal vacation combination based on holidays, efficiency, and personal preferences"},
	}
	c.JSON(http.StatusOK, strategies)
}

// Helper functions
func (h *Handler) getOrCreateYearConfig(year int) (models.YearConfig, error) {
	var config models.YearConfig
	var workWeekJSON string
	var optimizerNotes sql.NullString

	err := h.db.QueryRow(`SELECT id, year, vacation_days, COALESCE(reserved_days, 0), optimization_strategy, work_week, COALESCE(optimizer_notes, '') FROM year_config WHERE year = ?`, year).
		Scan(&config.ID, &config.Year, &config.VacationDays, &config.ReservedDays, &config.OptimizationStrategy, &workWeekJSON, &optimizerNotes)

	if err == sql.ErrNoRows {
		// Try to copy from previous year
		prevConfig, prevErr := h.getYearConfigOnly(year - 1)
		if prevErr == nil {
			config = prevConfig
			config.Year = year
		} else {
			// Use defaults
			config = models.YearConfig{
				Year:                 year,
				VacationDays:         22,
				ReservedDays:         0,
				OptimizationStrategy: models.StrategyBalanced,
				WorkWeek:             []string{"monday", "tuesday", "wednesday", "thursday", "friday"},
				OptimizerNotes:       "",
			}
		}

		workWeekJSON, _ := json.Marshal(config.WorkWeek)
		h.db.Exec(`INSERT INTO year_config (year, vacation_days, reserved_days, optimization_strategy, work_week, optimizer_notes) VALUES (?, ?, ?, ?, ?, ?)`,
			year, config.VacationDays, config.ReservedDays, config.OptimizationStrategy, string(workWeekJSON), config.OptimizerNotes)

		return config, nil
	}

	if err != nil {
		return config, err
	}

	json.Unmarshal([]byte(workWeekJSON), &config.WorkWeek)
	if optimizerNotes.Valid {
		config.OptimizerNotes = optimizerNotes.String
	}
	return config, nil
}

func (h *Handler) getYearConfigOnly(year int) (models.YearConfig, error) {
	var config models.YearConfig
	var workWeekJSON string
	var optimizerNotes sql.NullString

	err := h.db.QueryRow(`SELECT id, year, vacation_days, COALESCE(reserved_days, 0), optimization_strategy, work_week, COALESCE(optimizer_notes, '') FROM year_config WHERE year = ?`, year).
		Scan(&config.ID, &config.Year, &config.VacationDays, &config.ReservedDays, &config.OptimizationStrategy, &workWeekJSON, &optimizerNotes)

	if err != nil {
		return config, err
	}

	json.Unmarshal([]byte(workWeekJSON), &config.WorkWeek)
	if optimizerNotes.Valid {
		config.OptimizerNotes = optimizerNotes.String
	}
	return config, nil
}

func (h *Handler) getVacations(year int) ([]models.VacationDay, error) {
	rows, err := h.db.Query(`SELECT id, year, date, is_manual, COALESCE(note, '') FROM vacation_days WHERE year = ?`, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vacations []models.VacationDay
	for rows.Next() {
		var v models.VacationDay
		rows.Scan(&v.ID, &v.Year, &v.Date, &v.IsManual, &v.Note)
		vacations = append(vacations, v)
	}

	return vacations, nil
}

func (h *Handler) getOptimalVacations(year int) ([]models.OptimalVacation, error) {
	rows, err := h.db.Query(`SELECT id, year, date, block_id, consecutive_days FROM optimal_vacations WHERE year = ?`, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vacations []models.OptimalVacation
	for rows.Next() {
		var v models.OptimalVacation
		rows.Scan(&v.ID, &v.Year, &v.Date, &v.BlockID, &v.ConsecutiveDays)
		vacations = append(vacations, v)
	}

	return vacations, nil
}

func (h *Handler) buildCalendarDays(year int, config models.YearConfig, holidayList []holidays.PortugueseHoliday, manualVacations []models.VacationDay, optimalVacations []models.OptimalVacation) []models.CalendarDay {
	var days []models.CalendarDay

	// Create maps for quick lookup
	holidayMap := make(map[string]string)
	for _, h := range holidayList {
		holidayMap[h.Date] = h.Name
	}

	manualMap := make(map[string]bool)
	for _, v := range manualVacations {
		manualMap[v.Date] = true
	}

	optimalMap := make(map[string]int)
	for _, v := range optimalVacations {
		optimalMap[v.Date] = v.BlockID
	}

	workDaySet := make(map[string]bool)
	for _, d := range config.WorkWeek {
		workDaySet[d] = true
	}

	// Iterate through all days of the year
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		dayOfWeek := weekdayToString(d.Weekday())
		
		isWeekend := !workDaySet[dayOfWeek]
		holidayName, isHoliday := holidayMap[dateStr]
		isManual := manualMap[dateStr]
		blockID, isOptimal := optimalMap[dateStr]

		day := models.CalendarDay{
			Date:        dateStr,
			DayOfWeek:   dayOfWeek,
			IsWeekend:   isWeekend,
			IsHoliday:   isHoliday,
			HolidayName: holidayName,
			IsVacation:  isManual || isOptimal,
			IsManual:    isManual,
			IsOptimal:   isOptimal,
			BlockID:     blockID,
		}

		days = append(days, day)
	}

	return days
}

func (h *Handler) calculateSummary(totalVacation int, manualVacations []models.VacationDay, optimalVacations []models.OptimalVacation, holidayList []holidays.PortugueseHoliday) models.CalendarSummary {
	usedDays := len(manualVacations) + len(optimalVacations)
	
	// Calculate longest block
	blockDays := make(map[int]int)
	for _, v := range optimalVacations {
		if v.ConsecutiveDays > blockDays[v.BlockID] {
			blockDays[v.BlockID] = v.ConsecutiveDays
		}
	}
	
	longestBlock := 0
	for _, days := range blockDays {
		if days > longestBlock {
			longestBlock = days
		}
	}

	// Calculate total days off including bridged weekends
	// Collect all special days (vacations and holidays)
	specialDays := make(map[string]bool)
	for _, v := range manualVacations {
		specialDays[v.Date] = true
	}
	for _, v := range optimalVacations {
		specialDays[v.Date] = true
	}
	for _, h := range holidayList {
		specialDays[h.Date] = true
	}

	// Count weekends that are adjacent to special days (bridged)
	bridgedWeekends := 0
	countedWeekends := make(map[string]bool)
	
	for dateStr := range specialDays {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		
		// Check adjacent days for weekends
		for delta := -1; delta <= 1; delta += 2 { // -1 (before) and +1 (after)
			adjDate := date.AddDate(0, 0, delta)
			adjStr := adjDate.Format("2006-01-02")
			
			// If it's a weekend and not already counted
			if (adjDate.Weekday() == time.Saturday || adjDate.Weekday() == time.Sunday) && !countedWeekends[adjStr] {
				// Mark as counted and add to bridged count
				countedWeekends[adjStr] = true
				bridgedWeekends++
				
				// Also count the other weekend day if adjacent
				if adjDate.Weekday() == time.Saturday {
					sunday := adjDate.AddDate(0, 0, 1)
					sunStr := sunday.Format("2006-01-02")
					if !countedWeekends[sunStr] {
						countedWeekends[sunStr] = true
						bridgedWeekends++
					}
				} else if adjDate.Weekday() == time.Sunday {
					saturday := adjDate.AddDate(0, 0, -1)
					satStr := saturday.Format("2006-01-02")
					if !countedWeekends[satStr] {
						countedWeekends[satStr] = true
						bridgedWeekends++
					}
				}
			}
		}
	}

	return models.CalendarSummary{
		TotalVacationDays:     totalVacation,
		UsedVacationDays:      usedDays,
		RemainingVacationDays: totalVacation - usedDays,
		TotalHolidays:         len(holidayList),
		LongestVacationBlock:  longestBlock,
		TotalDaysOff:          usedDays + len(holidayList) + bridgedWeekends,
	}
}

func weekdayToString(day time.Weekday) string {
	switch day {
	case time.Monday:
		return "monday"
	case time.Tuesday:
		return "tuesday"
	case time.Wednesday:
		return "wednesday"
	case time.Thursday:
		return "thursday"
	case time.Friday:
		return "friday"
	case time.Saturday:
		return "saturday"
	case time.Sunday:
		return "sunday"
	}
	return ""
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
