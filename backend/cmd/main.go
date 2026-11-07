package main

import (
	"log"

	"booking-saas/config"
	"booking-saas/routes"
	"booking-saas/utils"
)

func main() {
	// Load Configuration
	cfg := config.LoadConfig()

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
	address := ":" + cfg.AppPort

	log.Printf("Booking SaaS API running at http://localhost%s", address)

	if err := router.Run(address); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
