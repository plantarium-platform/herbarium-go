package herbarium

import (
	"fmt"
	"github.com/plantarium-platform/herbarium-go/internal/manager"
	"github.com/plantarium-platform/herbarium-go/internal/storage"
	"github.com/plantarium-platform/herbarium-go/internal/storage/repos"
	"log"
)

func main() {
	// Set up in-memory storage
	stemStorage := &storage.HerbariumDB{} // Assuming HerbariumDB is an in-memory database for testing purposes

	// Initialize the repositories
	stemRepo := repos.NewStemRepository(stemStorage)
	mockLeafManager := new(manager.MockLeafManager)

	// Create the StemManager instance
	stemManager := manager.NewStemManager(stemRepo, mockLeafManager)

	// Create the PlatformManager instance
	platformManager := manager.NewPlatformManager(stemManager, mockLeafManager, "/path/to/your/base")

	// Start the platform
	err := platformManager.InitializePlatform()
	if err != nil {
		log.Fatalf("Failed to initialize the platform: %v", err)
	}

	fmt.Println("Platform started successfully")
}
