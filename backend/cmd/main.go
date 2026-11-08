package main

import (
	"log"
	"os"
	"strings"

	"booking-saas/config"
	"booking-saas/routes"
	"booking-saas/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load Configuration
	cfg := config.LoadConfig()
	mode := strings.TrimSpace(os.Getenv("GIN_MODE"))
	if mode == "" {
		mode = gin.ReleaseMode
	}
	gin.SetMode(mode)

	// Validate JWT Secret
	if cfg.JWT.Secret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	// Connect Database
	if err := utils.ConnectDatabase(cfg); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer utils.CloseDatabase()

	log.Println("Database connected successfully")

	// Setup Router
	router := routes.SetupRouter(cfg)

	// Start HTTP Server
	address := "127.0.0.1:" + cfg.AppPort

	log.Printf("Booking SaaS API running at http://%s", address)

	if err := router.Run(address); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
