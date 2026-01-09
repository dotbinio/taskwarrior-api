package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dotbinio/taskwarrior-api/internal/api"
	"github.com/dotbinio/taskwarrior-api/internal/auth"
	"github.com/dotbinio/taskwarrior-api/internal/config"
	"github.com/dotbinio/taskwarrior-api/internal/taskwarrior"
)

// @title           Taskwarrior API
// @version         1.0
// @description     Headless REST API for Taskwarrior

// @contact.name   API Support
// @contact.url    http://github.com/dotbinio/taskwarrior-api

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting Taskwarrior API server...")
	log.Printf("Data location: %s", cfg.Taskwarrior.DataLocation)
	log.Printf("Taskrc location: %s", cfg.Taskwarrior.TaskrcLocation)
	log.Printf("Server address: %s", cfg.GetAddress())

	// Initialize Taskwarrior client
	twClient := taskwarrior.NewClient(cfg.Taskwarrior.DataLocation, cfg.Taskwarrior.TaskrcLocation)

	// Initialize token validator
	validator := auth.NewTokenValidator(cfg.Auth.Tokens)

	// Setup router
	router := api.SetupRouter(cfg, twClient, validator)

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.GetAddress(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server listening on %s", cfg.GetAddress())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
