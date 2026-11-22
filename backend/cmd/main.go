package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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
	if err := cfg.ValidateSecurity(); err != nil {
		log.Fatal(err)
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

	server := &http.Server{Addr: address, Handler: router, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32 << 10}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	errs := make(chan error, 1)
	go func() { errs <- server.ListenAndServe() }()
	select {
	case err := <-errs:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			log.Printf("Server shutdown: %v", err)
			_ = server.Close()
		}
	}
}
