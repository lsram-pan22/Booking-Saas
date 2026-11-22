package main

import (
	"booking-saas/config"
	"booking-saas/migrations"
	"booking-saas/utils"
	"context"
	"log"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if err := utils.ConnectDatabase(config.LoadConfig()); err != nil {
		return err
	}
	defer utils.CloseDatabase()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	applied, err := migrations.Apply(ctx, utils.DB)
	if err != nil {
		return err
	}
	for _, name := range applied {
		log.Printf("applied public migration: %s", name)
	}
	log.Printf("migration complete: %d applied", len(applied))
	return nil
}
