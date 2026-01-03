package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"

	"github.com/bruno.lopes/calendar/backend/internal/holidays"
	"github.com/bruno.lopes/calendar/backend/internal/models"
)

// GitHubModel represents a model from GitHub Models API
type GitHubModel struct {
	Name         string `json:"name"`
	FriendlyName string `json:"friendly_name"`
	Publisher    string `json:"publisher"`
	Task         string `json:"task"`
}

// GetAvailableModels fetches available models from GitHub Models Catalog API
func (h *Handler) GetAvailableModels(c *gin.Context) {
	// Get API key from settings
	var apiKey string
	err := h.db.QueryRow("SELECT value FROM settings WHERE key = 'openai_api_key'").Scan(&apiKey)
	if err != nil || apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key not configured"})
		return
	}

	// Get AI provider
	var aiProvider string
	err = h.db.QueryRow("SELECT value FROM settings WHERE key = 'ai_provider'").Scan(&aiProvider)
	if err != nil || aiProvider == "" {
		aiProvider = "github"
	}

	if aiProvider == "openai" {
		// For OpenAI, fetch from OpenAI API
		client := openai.NewClient(apiKey)
		modelList, err := client.ListModels(context.Background())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch models: " + err.Error()})
			return
		}

		var chatModels []map[string]string
		for _, model := range modelList.Models {
			// Filter for chat models
			if strings.Contains(model.ID, "gpt") || strings.Contains(model.ID, "o1") || strings.Contains(model.ID, "o3") {
				chatModels = append(chatModels, map[string]string{
					"id":        model.ID,
					"name":      model.ID,
					"publisher": "OpenAI",
				})
			}
		}
		c.JSON(http.StatusOK, chatModels)
		return
	}

	// Fetch from GitHub Models Catalog API
	req, err := http.NewRequest("GET", "https://models.github.ai/catalog/models", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch models: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	if resp.StatusCode != 200 {
		c.JSON(resp.StatusCode, gin.H{"error": "GitHub API error: " + string(body)})
		return
	}

	// Parse the response
	var modelsResponse []map[string]interface{}
	if err := json.Unmarshal(body, &modelsResponse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse models"})
		return
	}

	// Filter for chat-capable models and format response
	var chatModels []map[string]string
	for _, model := range modelsResponse {
		// Get supported output modalities to filter for text generation models
		outputModalities, _ := model["supported_output_modalities"].([]interface{})
		hasTextOutput := false
		for _, m := range outputModalities {
			if m == "text" {
				hasTextOutput = true
				break
			}
		}

		// Skip embedding-only models
		name, _ := model["name"].(string)
		if strings.Contains(strings.ToLower(name), "embedding") {
			continue
		}

		if hasTextOutput {
			id, _ := model["id"].(string)
			publisher, _ := model["publisher"].(string)

			chatModels = append(chatModels, map[string]string{
				"id":        id,
				"name":      name,
				"publisher": publisher,
			})
		}
	}

	c.JSON(http.StatusOK, chatModels)
}

// Chat handles AI chat interactions
func (h *Handler) Chat(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	var input struct {
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get API key and provider from settings
	var apiKey string
	err = h.db.QueryRow("SELECT value FROM settings WHERE key = 'openai_api_key'").Scan(&apiKey)
	if err != nil || apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key not configured. Please set it in settings."})
		return
	}

	// Get AI provider setting (default to github for GitHub Copilot models)
	var aiProvider string
	err = h.db.QueryRow("SELECT value FROM settings WHERE key = 'ai_provider'").Scan(&aiProvider)
	if err != nil || aiProvider == "" {
		aiProvider = "github" // Default to GitHub Models
	}

	// Get selected model (default to openai/gpt-4o-mini)
	var selectedModel string
	err = h.db.QueryRow("SELECT value FROM settings WHERE key = 'ai_model'").Scan(&selectedModel)
	if err != nil || selectedModel == "" {
		selectedModel = "openai/gpt-4o-mini"
	}

	// Ensure model has publisher prefix for GitHub Models API
	if aiProvider == "github" && !strings.Contains(selectedModel, "/") {
		// Add openai/ prefix if no publisher specified
		selectedModel = "openai/" + selectedModel
	}

	// Save user message to history
	h.db.Exec(`INSERT INTO chat_history (year, role, content) VALUES (?, 'user', ?)`, year, input.Message)

	// Get calendar context
	calendarContext := h.getCalendarContext(year)

	// Get chat history for context
	chatHistory := h.getChatHistoryMessages(year, 10)

	// Create client based on provider
	var client *openai.Client

	switch aiProvider {
	case "github":
		// GitHub Models API (new endpoint)
		config := openai.DefaultConfig(apiKey)
		config.BaseURL = "https://models.github.ai/inference"
		client = openai.NewClientWithConfig(config)
	case "openai":
		client = openai.NewClient(apiKey)
	default:
		// Default to GitHub
		config := openai.DefaultConfig(apiKey)
		config.BaseURL = "https://models.github.ai/inference"
		client = openai.NewClientWithConfig(config)
	}

	// Build messages
	messages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: fmt.Sprintf(`You are a helpful vacation planning assistant. You help users plan their vacation days optimally around Portuguese holidays.

Current calendar context for year %d:
%s

You can help users:
1. Add or remove vacation days (both manual and optimized)
2. Change vacation settings (number of days, reserved days, optimization strategy, work week)
3. Run optimization to suggest best vacation days
4. Answer questions about holidays and vacation planning
5. Reorganize vacation days when all days are used

IMPORTANT RESTRICTION:
- You CANNOT set vacation days on holidays (national or municipal)
- Holidays are already days off, so users don't need to use vacation days for them
- If a user asks to set vacation on a holiday, politely explain that it's already a day off
- When suggesting vacation days, always check the holiday list and avoid those dates

IMPORTANT - Reserved days feature:
- Users can reserve some vacation days for emergencies/unexpected needs
- Reserved days are NOT used by the optimizer
- Available days for optimizer = Total vacation days - Reserved days - Manual days
- Help users understand that reserved days act as a buffer for last-minute needs

IMPORTANT - Handling vacation day limits:
- The context shows "Vacation days used" vs "Total available" and "Reserved"
- Manual vacation days are set directly by the user
- Optimized vacation days are calculated by the optimizer
- Reserved days are kept aside and not planned
- When all days are taken and user wants changes:
  * Suggest which existing days to remove to make room for new ones
  * Offer to clear optimized days and re-run optimization with new preferences
  * Suggest swapping days (remove some, add others)
  * Can increase total vacation days or reduce reserved days if user needs more

When reorganizing vacations:
- First remove the days that need to go, then add the new ones
- You can combine multiple actions: first a remove_vacation, then add_vacation
- If the user wants to completely reorganize, suggest: 1) clear all optimized days, 2) optionally clear manual days, 3) re-optimize

CRITICAL - Response format rules:
- DO NOT mention JSON, action blocks, or technical details to the user
- DO NOT say things like "Here's the action in JSON format" or "Executing action"
- Just naturally describe what you're doing: "I'll add those vacation days for you!" or "Done! I've cleared your vacations."
- Include the JSON action blocks in your response but don't reference them - the system processes them automatically
- Write responses as if you're a helpful assistant talking to a regular user, not a developer

Action formats (include these in your response but don't mention them to the user):
{"action": "add_vacation", "dates": ["2026-01-06", "2026-01-07"]}
{"action": "remove_vacation", "dates": ["2026-01-06"]}
{"action": "clear_optimized"}
{"action": "clear_all_vacations"}
{"action": "update_config", "vacation_days": 22, "reserved_days": 3, "optimization_strategy": "balanced", "work_week": ["monday","tuesday","wednesday","thursday","friday"]}
{"action": "optimize"}

You can chain multiple actions by including multiple JSON blocks in your response.

Available optimization strategies: 
- "bridge_holidays": Creates bridges between holidays and weekends for maximum connected time off
- "longest_blocks": Creates the longest possible consecutive vacation periods
- "balanced": Balance between efficiency (days off per vacation day) and block length

Available work week days: monday, tuesday, wednesday, thursday, friday, saturday, sunday`, year, calendarContext),
		},
	}

	// Add chat history
	for _, msg := range chatHistory {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Add current message
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: input.Message,
	})

	// Call AI API
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    selectedModel,
			Messages: messages,
		},
	)

	if err != nil {
		fmt.Printf("OpenAI API Error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get AI response: " + err.Error()})
		return
	}

	if len(resp.Choices) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No response from AI"})
		return
	}

	assistantMessage := resp.Choices[0].Message.Content

	// Save assistant message to history
	h.db.Exec(`INSERT INTO chat_history (year, role, content) VALUES (?, 'assistant', ?)`, year, assistantMessage)

	// Check for actions in the response
	action := h.parseAndExecuteAction(year, assistantMessage)

	c.JSON(http.StatusOK, gin.H{
		"message":    assistantMessage,
		"action":     action,
		"hasAction":  action != nil,
	})
}

// GetChatHistory returns chat history for a year
func (h *Handler) GetChatHistory(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	rows, err := h.db.Query(`SELECT id, year, role, content, created_at FROM chat_history WHERE year = ? ORDER BY created_at ASC`, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		rows.Scan(&msg.ID, &msg.Year, &msg.Role, &msg.Content, &msg.CreatedAt)
		messages = append(messages, msg)
	}

	c.JSON(http.StatusOK, messages)
}

// ClearChatHistory clears chat history for a year
func (h *Handler) ClearChatHistory(c *gin.Context) {
	yearStr := c.Param("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	_, err = h.db.Exec(`DELETE FROM chat_history WHERE year = ?`, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chat history cleared"})
}

// Helper functions
func (h *Handler) getCalendarContext(year int) string {
	config, _ := h.getOrCreateYearConfig(year)
	workCity := h.getWorkCity()
	holidayList := holidays.GetPortugueseHolidaysWithCity(year, workCity)
	manualVacations, _ := h.getVacations(year)
	optimalVacations, _ := h.getOptimalVacations(year)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Year: %d\n", year))
	sb.WriteString(fmt.Sprintf("Total vacation days available: %d\n", config.VacationDays))
	sb.WriteString(fmt.Sprintf("Reserved days (for emergencies): %d\n", config.ReservedDays))
	sb.WriteString(fmt.Sprintf("Optimization strategy: %s\n", config.OptimizationStrategy))
	sb.WriteString(fmt.Sprintf("Work week: %v\n", config.WorkWeek))
	if workCity != "" {
		sb.WriteString(fmt.Sprintf("Work city: %s (includes municipal holidays)\n", workCity))
	}
	
	sb.WriteString("\nPortuguese Holidays:\n")
	for _, h := range holidayList {
		sb.WriteString(fmt.Sprintf("- %s: %s (%s)\n", h.Date, h.Name, h.Type))
	}

	manualCount := len(manualVacations)
	optimizedCount := len(optimalVacations)
	usedDays := manualCount + optimizedCount
	availableForPlanning := config.VacationDays - config.ReservedDays
	remaining := availableForPlanning - usedDays

	if manualCount > 0 {
		sb.WriteString(fmt.Sprintf("\nManually set vacation days (%d days):\n", manualCount))
		for _, v := range manualVacations {
			sb.WriteString(fmt.Sprintf("- %s\n", v.Date))
		}
	} else {
		sb.WriteString("\nNo manual vacation days set.\n")
	}

	if optimizedCount > 0 {
		sb.WriteString(fmt.Sprintf("\nOptimized vacation days (%d days):\n", optimizedCount))
		// Group by block
		blocks := make(map[int][]string)
		for _, v := range optimalVacations {
			blocks[v.BlockID] = append(blocks[v.BlockID], v.Date)
		}
		for blockID, dates := range blocks {
			sb.WriteString(fmt.Sprintf("  Block %d: %s to %s (%d days)\n", blockID, dates[0], dates[len(dates)-1], len(dates)))
		}
	} else {
		sb.WriteString("\nNo optimized vacation days. Run optimization to get suggestions.\n")
	}

	sb.WriteString(fmt.Sprintf("\n=== VACATION BUDGET ===\n"))
	sb.WriteString(fmt.Sprintf("Total vacation days: %d\n", config.VacationDays))
	sb.WriteString(fmt.Sprintf("Reserved for emergencies: %d\n", config.ReservedDays))
	sb.WriteString(fmt.Sprintf("Available for planning: %d\n", availableForPlanning))
	sb.WriteString(fmt.Sprintf("Manual days used: %d\n", manualCount))
	sb.WriteString(fmt.Sprintf("Optimized days used: %d\n", optimizedCount))
	sb.WriteString(fmt.Sprintf("Total planned: %d days\n", usedDays))
	sb.WriteString(fmt.Sprintf("Remaining to plan: %d days\n", remaining))
	
	if remaining < 0 {
		sb.WriteString(fmt.Sprintf("⚠️ OVER BUDGET by %d days! Need to remove some vacation days or increase total.\n", -remaining))
	} else if remaining == 0 {
		sb.WriteString("✓ All plannable vacation days are allocated.\n")
	}

	return sb.String()
}

func (h *Handler) getChatHistoryMessages(year int, limit int) []openai.ChatCompletionMessage {
	rows, err := h.db.Query(`SELECT role, content FROM chat_history WHERE year = ? ORDER BY created_at DESC LIMIT ?`, year, limit)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var messages []openai.ChatCompletionMessage
	for rows.Next() {
		var role, content string
		rows.Scan(&role, &content)
		messages = append([]openai.ChatCompletionMessage{{Role: role, Content: content}}, messages...)
	}

	return messages
}

func (h *Handler) parseAndExecuteAction(year int, message string) map[string]interface{} {
	// Find all JSON action blocks in the message
	var allActions []map[string]interface{}
	searchStart := 0
	
	for {
		start := strings.Index(message[searchStart:], "{\"action\"")
		if start == -1 {
			break
		}
		start += searchStart

		// Find matching closing brace
		depth := 0
		end := -1
		for i := start; i < len(message); i++ {
			if message[i] == '{' {
				depth++
			} else if message[i] == '}' {
				depth--
				if depth == 0 {
					end = i + 1
					break
				}
			}
		}

		if end == -1 {
			break
		}

		jsonStr := message[start:end]
		var action map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &action); err == nil {
			// Execute this action
			h.executeSingleAction(year, action)
			allActions = append(allActions, action)
		}
		
		searchStart = end
	}

	if len(allActions) == 0 {
		return nil
	}

	// Return the combined result
	if len(allActions) == 1 {
		return allActions[0]
	}

	// Return info about multiple actions
	return map[string]interface{}{
		"action":       "multiple",
		"actions":      allActions,
		"actionCount":  len(allActions),
	}
}

func (h *Handler) executeSingleAction(year int, action map[string]interface{}) {
	actionType, ok := action["action"].(string)
	if !ok {
		return
	}

	// Get holidays for this year to validate vacation dates
	workCity := h.getWorkCity()
	holidayList := holidays.GetPortugueseHolidaysWithCity(year, workCity)
	holidayDates := make(map[string]bool)
	for _, hol := range holidayList {
		holidayDates[hol.Date] = true
	}

	switch actionType {
	case "add_vacation":
		if dates, ok := action["dates"].([]interface{}); ok {
			var skippedHolidays []string
			for _, d := range dates {
				if dateStr, ok := d.(string); ok {
					// Skip if the date is a holiday
					if holidayDates[dateStr] {
						skippedHolidays = append(skippedHolidays, dateStr)
						continue
					}
					h.db.Exec(`INSERT OR REPLACE INTO vacation_days (year, date, is_manual) VALUES (?, ?, TRUE)`, year, dateStr)
				}
			}
			if len(skippedHolidays) > 0 {
				action["skipped_holidays"] = skippedHolidays
			}
		}
	case "remove_vacation":
		if dates, ok := action["dates"].([]interface{}); ok {
			for _, d := range dates {
				if dateStr, ok := d.(string); ok {
					// Remove from both manual and optimized tables
					h.db.Exec(`DELETE FROM vacation_days WHERE year = ? AND date = ?`, year, dateStr)
					h.db.Exec(`DELETE FROM optimal_vacations WHERE year = ? AND date = ?`, year, dateStr)
				}
			}
		}
	case "clear_optimized":
		// Clear only optimized vacation days, keep manual ones
		h.db.Exec(`DELETE FROM optimal_vacations WHERE year = ?`, year)
		action["cleared"] = "optimized"
	case "clear_all_vacations":
		// Clear both manual and optimized vacation days
		h.db.Exec(`DELETE FROM vacation_days WHERE year = ?`, year)
		h.db.Exec(`DELETE FROM optimal_vacations WHERE year = ?`, year)
		action["cleared"] = "all"
	case "update_config":
		updates := make(map[string]interface{})
		if vacDays, ok := action["vacation_days"].(float64); ok {
			updates["vacation_days"] = int(vacDays)
		}
		if reservedDays, ok := action["reserved_days"].(float64); ok {
			updates["reserved_days"] = int(reservedDays)
		}
		if strategy, ok := action["optimization_strategy"].(string); ok {
			updates["optimization_strategy"] = strategy
		}
		if workWeek, ok := action["work_week"].([]interface{}); ok {
			var days []string
			for _, d := range workWeek {
				if dayStr, ok := d.(string); ok {
					days = append(days, dayStr)
				}
			}
			workWeekJSON, _ := json.Marshal(days)
			updates["work_week"] = string(workWeekJSON)
		}

		if len(updates) > 0 {
			for key, value := range updates {
				h.db.Exec(fmt.Sprintf(`UPDATE year_config SET %s = ?, updated_at = CURRENT_TIMESTAMP WHERE year = ?`, key), value, year)
			}
		}
	case "optimize":
		// Trigger optimization - this will be handled by frontend calling the optimize endpoint
		action["triggerOptimize"] = true
	}
}
