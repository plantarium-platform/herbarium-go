package haproxy

import (
	"fmt"
	"log"
)

// HAProxyClientInterface defines the contract for HAProxy client interactions.
type HAProxyClientInterface interface {
	BindStem(backendName string) error
	BindLeaf(backendName, leafID, serviceAddress string, servicePort int) error
	UnbindLeaf(backendName, haProxyServer string) error
	ReplaceLeaf(backendName, oldHAProxyServer, newHAProxyServer, serviceAddress string, servicePort int) error
	UnbindStem(backendName string) error
}

// HAProxyConfig represents the HAProxy configuration needed for initialization.
type HAProxyConfig struct {
	APIURL   string
	Username string
	Password string
}

// HAProxyClient provides a high-level interface for managing the HAProxy configuration.
type HAProxyClient struct {
	configManager         HAProxyConfigurationManagerInterface // Using the interface here
	transactionMiddleware TransactionMiddleware
}

// NewHAProxyClient initializes and returns an HAProxyClient that implements HAProxyClientInterface.
func NewHAProxyClient(config HAProxyConfig, configManager HAProxyConfigurationManagerInterface) HAProxyClientInterface {
	transactionMiddleware := NewTransactionMiddleware(configManager)

	// Return the client with the necessary configurations
	return &HAProxyClient{
		configManager:         configManager,
		transactionMiddleware: transactionMiddleware,
	}
}

// BindStem creates a backend for a stem in HAProxy.
func (c *HAProxyClient) BindStem(backendName string) error {
	log.Printf("[HAProxyClient] Attempting to bind stem as backend: %s", backendName)
	return c.transactionMiddleware(func(transactionID string) error {
		log.Printf("[HAProxyClient] Starting transaction for backend creation: transactionID=%s, backendName=%s", transactionID, backendName)

		// Create the backend for the stem if it doesn't exist
		err := c.configManager.CreateBackend(backendName, transactionID)
		if err != nil {
			log.Printf("[HAProxyClient] Failed to create backend: backendName=%s, transactionID=%s, error=%v", backendName, transactionID, err)
			return fmt.Errorf("failed to create backend: %v", err)
		}

		log.Printf("[HAProxyClient] Successfully created backend: backendName=%s, transactionID=%s", backendName, transactionID)
		return nil
	})()
}

// BindLeaf adds a leaf service to the specified backend using HAProxy server details.
func (c *HAProxyClient) BindLeaf(backendName, leafID, serviceAddress string, servicePort int) error {
	log.Printf("Binding leaf: Backend=%s, LeafID=%s, Address=%s:%d", backendName, leafID, serviceAddress, servicePort)

	return c.transactionMiddleware(func(transactionID string) error {
		log.Printf("Starting HAProxy transaction for binding: TransactionID=%s", transactionID)

		// Construct service address as IP + port
		address := fmt.Sprintf("%s:%d", serviceAddress, servicePort)

		// Add the leaf as a service in the backend using leaf ID and service address
		err := c.configManager.AddServer(backendName, leafID, serviceAddress, servicePort, transactionID)
		if err != nil {
			log.Printf("Failed to add server to HAProxy: Backend=%s, LeafID=%s, Address=%s, TransactionID=%s, Error=%v", backendName, leafID, address, transactionID, err)
			return fmt.Errorf("failed to bind leaf service: %v", err)
		}

		log.Printf("Successfully bound leaf: Backend=%s, LeafID=%s, Address=%s, TransactionID=%s", backendName, leafID, address, transactionID)
		return nil
	})()
}

// UnbindLeaf removes a leaf service from the specified backend using HAProxy server details.
func (c *HAProxyClient) UnbindLeaf(backendName, haProxyServer string) error {
	return c.transactionMiddleware(func(transactionID string) error {
		// Remove the leaf service from the backend
		err := c.configManager.DeleteServer(backendName, haProxyServer, transactionID)
		if err != nil {
			return fmt.Errorf("failed to unbind leaf service: %v", err)
		}
		return nil
	})()
}

// ReplaceLeaf replaces an existing leaf service with a new one by using the HAProxy server name.
func (c *HAProxyClient) ReplaceLeaf(backendName, oldHAProxyServer, newHAProxyServer, serviceAddress string, servicePort int) error {
	return c.transactionMiddleware(func(transactionID string) error {
		// Remove the old leaf service
		err := c.configManager.DeleteServer(backendName, oldHAProxyServer, transactionID)
		if err != nil {
			return fmt.Errorf("failed to remove old leaf service: %v", err)
		}

		// Add the new leaf service with separate address and port
		err = c.configManager.AddServer(backendName, newHAProxyServer, serviceAddress, servicePort, transactionID)
		if err != nil {
			return fmt.Errorf("failed to add new leaf service: %v", err)
		}

		return nil
	})()
}

// UnbindStem removes the backend for the stem from HAProxy.
func (c *HAProxyClient) UnbindStem(backendName string) error {
	return c.transactionMiddleware(func(transactionID string) error {
		// Delete the backend for the stem
		err := c.configManager.DeleteServer(backendName, "", transactionID) // Deletes all services under the backend
		if err != nil {
			return fmt.Errorf("failed to remove backend: %v", err)
		}
		return nil
	})()
}
