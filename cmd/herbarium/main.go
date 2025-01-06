package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/plantarium-platform/herbarium-go/internal/manager"
)

func main() {
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

	log.Println("Platform started successfully")
	log.Println("Waiting for termination signal...")

	// Create a channel to listen for OS signals
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	// Block until a termination signal is received
	<-signalChannel

	log.Println("Termination signal received. Shutting down...")
	// Perform any necessary cleanup here before exiting
}
