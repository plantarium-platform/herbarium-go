package main

import (
	"fmt"
	"log"

	"github.com/plantarium-platform/herbarium-go/internal/manager"
	"go.uber.org/zap"
)

func main() {
	// Initialize the Zap logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize Zap logger: %v", err)
	}
	defer logger.Sync()

	// Replace the standard logger with Zap
	zap.ReplaceGlobals(logger)
	if err := zap.RedirectStdLog(logger); err != nil {
		log.Fatalf("Failed to redirect standard log to Zap: %v", err)
	}

	// Create a new PlatformManager instance with dependencies initialized internally
	platformManager, err := manager.NewPlatformManagerWithDI()
	if err != nil {
		log.Fatalf("Failed to create platform manager: %v", err)
	}

	// Start the platform
	err = platformManager.InitializePlatform()
	if err != nil {
		log.Fatalf("Failed to initialize the platform: %v", err)
	}

	fmt.Println("Platform started successfully")
}
