package routes

import (
	"booking-saas/config"
	"booking-saas/utils"

	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api")

	// Health Check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
			"message": "Booking SaaS API is running",
		})
	})

	// Feature Routes
	AuthRoutes(api, utils.DB, cfg)

	return router
}
