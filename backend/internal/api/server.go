package api

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/bruno.lopes/calendar/backend/internal/api/handlers"
)

// Version is set at build time
var Version = "dev"

type Server struct {
	db     *sql.DB
	router *gin.Engine
}

func NewServer(db *sql.DB) *Server {
	s := &Server{
		db:     db,
		router: gin.Default(),
	}

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	s.router.Use(cors.New(config))

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	h := handlers.NewHandler(s.db)

	api := s.router.Group("/api")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Version endpoint
		api.GET("/version", func(c *gin.Context) {
			version := Version
			if v := os.Getenv("APP_VERSION"); v != "" {
				version = v
			}
			c.JSON(http.StatusOK, gin.H{"version": version})
		})

		// Calendar endpoints
		api.GET("/calendar/:year", h.GetCalendar)
		api.POST("/calendar/:year/optimize", h.OptimizeVacations)
		api.DELETE("/calendar/:year/optimized", h.ClearOptimizedVacations)
		api.GET("/calendar/:year/suggestions", h.GetVacationSuggestions)

		// Vacation days endpoints
		api.GET("/vacations/:year", h.GetVacations)
		api.POST("/vacations/:year", h.AddVacation)
		api.DELETE("/vacations/:year/:date", h.RemoveVacation)
		api.PUT("/vacations/:year/bulk", h.BulkUpdateVacations)

		// Holidays endpoints
		api.GET("/holidays/:year", h.GetHolidays)
		api.GET("/holidays/:year/status", h.GetHolidayStatus)
		api.GET("/holidays/status", h.GetAllHolidayStatuses)
		api.POST("/holidays/:year/refresh", h.RefreshHolidays)
		api.GET("/cities", h.GetAvailableCities)

		// Year config endpoints
		api.GET("/config/:year", h.GetYearConfig)
		api.PUT("/config/:year", h.UpdateYearConfig)
		api.POST("/config/:year/copy-from/:sourceYear", h.CopyYearConfig)

		// Settings endpoints
		api.GET("/settings", h.GetSettings)
		api.PUT("/settings", h.UpdateSettings)
		api.GET("/settings/:key", h.GetSetting)
		api.PUT("/settings/:key", h.UpdateSetting)

		// Chat endpoints
		api.POST("/chat/:year", h.Chat)
		api.GET("/chat/:year/history", h.GetChatHistory)
		api.DELETE("/chat/:year/history", h.ClearChatHistory)

		// AI models endpoint
		api.GET("/models", h.GetAvailableModels)

		// Work week presets
		api.GET("/presets/work-week", h.GetWorkWeekPresets)
		api.GET("/presets/strategies", h.GetOptimizationStrategies)
	}
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}
