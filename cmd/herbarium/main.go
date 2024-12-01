package main

import (
	"fmt"
	"log"

	"github.com/plantarium-platform/herbarium-go/internal/manager"
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
)

func main() {
	// Initialize in-memory storage
	stemStorage := storage.GetHerbariumDB() // Singleton HerbariumDB instance

	// Initialize repositories
	stemRepo := repos.NewStemRepository(stemStorage)
	leafRepo := repos.NewLeafRepository(stemStorage)

	// Create MockHAProxyClient (assuming it lives in the same package as mocks)
	mockHAProxyClient := new(manager.MockHAProxyClient)

	// Create the LeafManager
	leafManager := manager.NewLeafManager(leafRepo, mockHAProxyClient, stemRepo)

	// Create the StemManager
	stemManager := manager.NewStemManager(stemRepo, leafManager, mockHAProxyClient)

	// Create the PlatformManager
	platformManager := manager.NewPlatformManager(stemManager, leafManager, "/path/to/your/base")

	// Start the platform
	err := platformManager.InitializePlatform()
	if err != nil {
		log.Fatalf("Failed to initialize the platform: %v", err)
	}

	fmt.Println("Platform started successfully")
}
