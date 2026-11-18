package routes

import (
	"time"

	"booking-saas/config"
	"booking-saas/middleware"
	"booking-saas/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.New()
	// Client-IP limits trust forwarding headers only from configured peers.
	_ = router.SetTrustedProxies(cfg.TrustedProxies)
	router.Use(middleware.SecurityHeaders(), middleware.SecurityAudit(), middleware.SafeRecovery(), middleware.LimitRequestBody())
	origins := cfg.CORSOrigins
	if len(origins) == 0 {
		origins = []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	}

	// CORS Middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins: origins,
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
	api.Use(middleware.AuthenticationLimits())

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
