package main

import (
	"fmt"
	"log"

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

	fmt.Println("Platform started successfully")
}
