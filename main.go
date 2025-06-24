package main

import (
	"context"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"luxsuv-backend/handlers"
	"luxsuv-backend/logger"
	"luxsuv-backend/repository"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Initialize logger
	logger := logger.NewLogger()

	// Debug: Check environment variables
	connString := os.Getenv("DATABASE_URL")
	logger.Info(fmt.Sprintf("DATABASE_URL: %s", connString))
	if connString == "" {
		logger.Error("DATABASE_URL environment variable not set")
		return
	}

	// Initialize database
	ctx := context.Background()
	repo, err := repository.NewBookingRepository(ctx, connString)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to initialize repository: %v", err))
		return
	}
	defer repo.Close()

	// Set up routers
	riderRouter := handlers.SetupRiderRouter(repo)
	driverRouter := handlers.SetupDriverRouter(repo)

	// Mount routers
	mux := http.NewServeMux()
	mux.Handle("/rider/", http.StripPrefix("/rider", riderRouter))
	mux.Handle("/driver/", http.StripPrefix("/driver", driverRouter))

	// Create HTTP server with graceful shutdown
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(fmt.Sprintf("Server failed: %v", err))
			os.Exit(1)
		}
	}()
	logger.Info("Server started on :8080")

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error(fmt.Sprintf("Server shutdown failed: %v", err))
	}
	logger.Info("Server stopped")
}
