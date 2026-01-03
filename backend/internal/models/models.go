package models

import "time"

// Settings represents application settings
type Settings struct {
	ID        int64     `json:"id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// YearConfig represents configuration for a specific year
type YearConfig struct {
	ID                   int64    `json:"id"`
	Year                 int      `json:"year"`
	VacationDays         int      `json:"vacation_days"`
	ReservedDays         int      `json:"reserved_days"`
	OptimizationStrategy string   `json:"optimization_strategy"`
	WorkWeek             []string `json:"work_week"`
	OptimizerNotes       string   `json:"optimizer_notes"`
	CreatedAt            string   `json:"created_at"`
	UpdatedAt            string   `json:"updated_at"`
}

// VacationDay represents a vacation day
type VacationDay struct {
	ID        int64  `json:"id"`
	Year      int    `json:"year"`
	Date      string `json:"date"`
	IsManual  bool   `json:"is_manual"`
	Note      string `json:"note,omitempty"`
	CreatedAt string `json:"created_at"`
}

// OptimalVacation represents a calculated optimal vacation day
type OptimalVacation struct {
	ID              int64  `json:"id"`
	Year            int    `json:"year"`
	Date            string `json:"date"`
	BlockID         int    `json:"block_id"`
	ConsecutiveDays int    `json:"consecutive_days"`
	CreatedAt       string `json:"created_at"`
}

// Holiday represents a Portuguese holiday
type Holiday struct {
	ID   int64  `json:"id"`
	Year int    `json:"year"`
	Date string `json:"date"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ChatMessage represents a message in the chat history
type ChatMessage struct {
	ID        int64  `json:"id"`
	Year      int    `json:"year"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// VacationBlock represents a block of consecutive vacation days
type VacationBlock struct {
	StartDate       string   `json:"start_date"`
	EndDate         string   `json:"end_date"`
	TotalDays       int      `json:"total_days"`
	VacationDaysUsed int     `json:"vacation_days_used"`
	Dates           []string `json:"dates"`
	Holidays        []string `json:"holidays"`
	Weekends        []string `json:"weekends"`
}

// CalendarDay represents a single day in the calendar
type CalendarDay struct {
	Date        string `json:"date"`
	DayOfWeek   string `json:"day_of_week"`
	IsWeekend   bool   `json:"is_weekend"`
	IsHoliday   bool   `json:"is_holiday"`
	HolidayName string `json:"holiday_name,omitempty"`
	IsVacation  bool   `json:"is_vacation"`
	IsManual    bool   `json:"is_manual"`
	IsOptimal   bool   `json:"is_optimal"`
	BlockID     int    `json:"block_id,omitempty"`
}

// CalendarResponse represents the full calendar data for a year
type CalendarResponse struct {
	Year             int             `json:"year"`
	Config           YearConfig      `json:"config"`
	Days             []CalendarDay   `json:"days"`
	Holidays         []Holiday       `json:"holidays"`
	VacationBlocks   []VacationBlock `json:"vacation_blocks"`
	ManualVacations  []VacationDay   `json:"manual_vacations"`
	OptimalVacations []OptimalVacation `json:"optimal_vacations"`
	Summary          CalendarSummary `json:"summary"`
}

// CalendarSummary provides statistics about the calendar
type CalendarSummary struct {
	TotalVacationDays    int `json:"total_vacation_days"`
	UsedVacationDays     int `json:"used_vacation_days"`
	RemainingVacationDays int `json:"remaining_vacation_days"`
	TotalHolidays        int `json:"total_holidays"`
	LongestVacationBlock int `json:"longest_vacation_block"`
	TotalDaysOff         int `json:"total_days_off"`
}

// OptimizationStrategy constants
const (
	StrategyBridgeHolidays = "bridge_holidays"
	StrategyLongestBlocks  = "longest_blocks"
	StrategyBalanced       = "balanced"
	StrategySmart          = "smart"
)

// WorkWeek days
var AllWeekDays = []string{
	"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday",
}

// WorkWeekPresets represents preset work week configurations
var WorkWeekPresets = map[string][]string{
	"standard":     {"monday", "tuesday", "wednesday", "thursday", "friday"},
	"four_day":     {"monday", "tuesday", "wednesday", "thursday"},
	"four_day_fri": {"tuesday", "wednesday", "thursday", "friday"},
	"six_day":      {"monday", "tuesday", "wednesday", "thursday", "friday", "saturday"},
	"custom":       {},
}
