package routes

import (
	"time"

	"booking-saas/config"
	"booking-saas/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()
	// The API listens directly on localhost without a reverse proxy.
	_ = router.SetTrustedProxies(nil)

	// CORS Middleware
	corsOrigins := cfg.CORSOrigins
	if len(corsOrigins) == 0 {
		corsOrigins = []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins: corsOrigins,
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
		MaxAge: 12 * time.Hour,
	}))

	api := router.Group("/api/v1")

	// Health Check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
			"message": "Booking SaaS API is running",
		})
	})

	// Feature Routes
	SuperAdminRoutes(api, utils.DB, cfg)
	TenantRoutes(api, utils.DB, cfg)

	return router
}
