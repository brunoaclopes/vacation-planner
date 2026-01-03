package optimizer

import (
	"sort"
	"time"

	"github.com/bruno.lopes/calendar/backend/internal/holidays"
	"github.com/bruno.lopes/calendar/backend/internal/models"
)

// Optimizer handles vacation optimization
type Optimizer struct {
	Year                 int
	VacationDays         int
	WorkWeek             []string
	Strategy             string
	Holidays             []holidays.PortugueseHoliday
	ManualVacations      []string
}

// NewOptimizer creates a new optimizer
func NewOptimizer(year, vacationDays int, workWeek []string, strategy string) *Optimizer {
	return NewOptimizerWithCity(year, vacationDays, workWeek, strategy, "")
}

// NewOptimizerWithCity creates a new optimizer with city-specific holidays
func NewOptimizerWithCity(year, vacationDays int, workWeek []string, strategy, city string) *Optimizer {
	return &Optimizer{
		Year:         year,
		VacationDays: vacationDays,
		WorkWeek:     workWeek,
		Strategy:     strategy,
		Holidays:     holidays.GetPortugueseHolidaysWithCity(year, city),
	}
}

// SetManualVacations sets manually chosen vacation days
func (o *Optimizer) SetManualVacations(vacations []string) {
	o.ManualVacations = vacations
}

// Optimize calculates optimal vacation days based on strategy
func (o *Optimizer) Optimize() []models.VacationBlock {
	switch o.Strategy {
	case models.StrategyBridgeHolidays:
		return o.bridgeHolidays()
	case models.StrategyLongestBlocks:
		return o.longestBlocks()
	case models.StrategyBalanced:
		return o.balanced()
	default:
		return o.balanced()
	}
}

// bridgeHolidays focuses on creating bridges between holidays and weekends
func (o *Optimizer) bridgeHolidays() []models.VacationBlock {
	opportunities := o.findBridgeOpportunities()
	
	// Sort by efficiency (days off gained per vacation day used)
	sort.Slice(opportunities, func(i, j int) bool {
		effI := float64(opportunities[i].TotalDays) / float64(opportunities[i].VacationDaysUsed)
		effJ := float64(opportunities[j].TotalDays) / float64(opportunities[j].VacationDaysUsed)
		return effI > effJ
	})

	return o.selectBlocks(opportunities)
}

// longestBlocks focuses on creating the longest possible vacation blocks
func (o *Optimizer) longestBlocks() []models.VacationBlock {
	opportunities := o.findAllOpportunities()
	
	// Sort by total consecutive days
	sort.Slice(opportunities, func(i, j int) bool {
		return opportunities[i].TotalDays > opportunities[j].TotalDays
	})

	return o.selectBlocks(opportunities)
}

// balanced combines both strategies
func (o *Optimizer) balanced() []models.VacationBlock {
	opportunities := o.findAllOpportunities()
	
	// Score based on both efficiency and total days
	sort.Slice(opportunities, func(i, j int) bool {
		effI := float64(opportunities[i].TotalDays) / float64(opportunities[i].VacationDaysUsed)
		effJ := float64(opportunities[j].TotalDays) / float64(opportunities[j].VacationDaysUsed)
		
		// Weight: 60% efficiency, 40% total days
		scoreI := effI*0.6 + float64(opportunities[i].TotalDays)*0.4
		scoreJ := effJ*0.6 + float64(opportunities[j].TotalDays)*0.4
		
		return scoreI > scoreJ
	})

	return o.selectBlocks(opportunities)
}

// findBridgeOpportunities finds opportunities to bridge holidays with weekends
func (o *Optimizer) findBridgeOpportunities() []models.VacationBlock {
	var opportunities []models.VacationBlock
	
	for _, holiday := range o.Holidays {
		holidayDate, _ := time.Parse("2006-01-02", holiday.Date)
		dayOfWeek := holidayDate.Weekday()
		
		// Check for bridge opportunities based on day of week
		switch dayOfWeek {
		case time.Monday:
			// Friday before could create 4-day weekend
			block := o.calculateBlock(holidayDate.AddDate(0, 0, -3), holidayDate)
			if block.VacationDaysUsed > 0 {
				opportunities = append(opportunities, block)
			}
		case time.Tuesday:
			// Monday before creates bridge
			block := o.calculateBlock(holidayDate.AddDate(0, 0, -1), holidayDate)
			if block.VacationDaysUsed > 0 {
				opportunities = append(opportunities, block)
			}
		case time.Thursday:
			// Friday after creates bridge
			block := o.calculateBlock(holidayDate, holidayDate.AddDate(0, 0, 1))
			if block.VacationDaysUsed > 0 {
				opportunities = append(opportunities, block)
			}
		case time.Friday:
			// Monday after could extend weekend
			block := o.calculateBlock(holidayDate, holidayDate.AddDate(0, 0, 3))
			if block.VacationDaysUsed > 0 {
				opportunities = append(opportunities, block)
			}
		case time.Wednesday:
			// Could take Mon-Tue or Thu-Fri for longer break
			block1 := o.calculateBlock(holidayDate.AddDate(0, 0, -2), holidayDate)
			block2 := o.calculateBlock(holidayDate, holidayDate.AddDate(0, 0, 2))
			if block1.VacationDaysUsed > 0 {
				opportunities = append(opportunities, block1)
			}
			if block2.VacationDaysUsed > 0 {
				opportunities = append(opportunities, block2)
			}
		}
	}
	
	return opportunities
}

// findAllOpportunities finds all possible vacation opportunities
func (o *Optimizer) findAllOpportunities() []models.VacationBlock {
	opportunities := o.findBridgeOpportunities()
	
	// Also look for week-long opportunities around holidays
	for _, holiday := range o.Holidays {
		holidayDate, _ := time.Parse("2006-01-02", holiday.Date)
		
		// Try full week around holiday
		weekStart := o.findWeekStart(holidayDate)
		weekEnd := weekStart.AddDate(0, 0, 6)
		block := o.calculateBlock(weekStart, weekEnd)
		if block.VacationDaysUsed > 0 && block.TotalDays >= 7 {
			opportunities = append(opportunities, block)
		}
		
		// Try two weeks around holiday
		twoWeekStart := weekStart.AddDate(0, 0, -7)
		block2 := o.calculateBlock(twoWeekStart, weekEnd)
		if block2.VacationDaysUsed > 0 && block2.TotalDays >= 14 {
			opportunities = append(opportunities, block2)
		}
	}
	
	return o.deduplicateBlocks(opportunities)
}

// calculateBlock calculates vacation block details between two dates
func (o *Optimizer) calculateBlock(start, end time.Time) models.VacationBlock {
	block := models.VacationBlock{
		StartDate: start.Format("2006-01-02"),
		EndDate:   end.Format("2006-01-02"),
	}
	
	current := start
	for !current.After(end) {
		dateStr := current.Format("2006-01-02")
		block.Dates = append(block.Dates, dateStr)
		block.TotalDays++
		
		if o.isWeekend(current) {
			block.Weekends = append(block.Weekends, dateStr)
		} else if isHol, _ := holidays.IsHoliday(current, o.Holidays); isHol {
			block.Holidays = append(block.Holidays, dateStr)
		} else if o.isWorkDay(current) && !o.isManualVacation(dateStr) {
			block.VacationDaysUsed++
		}
		
		current = current.AddDate(0, 0, 1)
	}
	
	return block
}

// selectBlocks selects vacation blocks within available vacation days
func (o *Optimizer) selectBlocks(opportunities []models.VacationBlock) []models.VacationBlock {
	var selected []models.VacationBlock
	usedDays := 0 // Start from 0 since VacationDays already accounts for manual/reserved
	usedDates := make(map[string]bool)
	
	// Mark manual vacation dates as used to prevent overlap
	for _, v := range o.ManualVacations {
		usedDates[v] = true
	}
	
	for _, block := range opportunities {
		// Check if we have enough days left
		if usedDays+block.VacationDaysUsed > o.VacationDays {
			continue
		}
		
		// Check for overlapping dates
		hasOverlap := false
		for _, date := range block.Dates {
			if usedDates[date] {
				hasOverlap = true
				break
			}
		}
		
		if hasOverlap {
			continue
		}
		
		// Add block
		selected = append(selected, block)
		usedDays += block.VacationDaysUsed
		for _, date := range block.Dates {
			usedDates[date] = true
		}
		
		if usedDays >= o.VacationDays {
			break
		}
	}
	
	return selected
}

// Helper functions
func (o *Optimizer) isWeekend(date time.Time) bool {
	day := date.Weekday()
	dayName := weekdayToString(day)
	
	for _, workDay := range o.WorkWeek {
		if workDay == dayName {
			return false
		}
	}
	return true
}

func (o *Optimizer) isWorkDay(date time.Time) bool {
	return !o.isWeekend(date)
}

func (o *Optimizer) isManualVacation(date string) bool {
	for _, v := range o.ManualVacations {
		if v == date {
			return true
		}
	}
	return false
}

func (o *Optimizer) findWeekStart(date time.Time) time.Time {
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, -1)
	}
	return date
}

func (o *Optimizer) deduplicateBlocks(blocks []models.VacationBlock) []models.VacationBlock {
	seen := make(map[string]bool)
	var unique []models.VacationBlock
	
	for _, block := range blocks {
		key := block.StartDate + "-" + block.EndDate
		if !seen[key] {
			seen[key] = true
			unique = append(unique, block)
		}
	}
	
	return unique
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
