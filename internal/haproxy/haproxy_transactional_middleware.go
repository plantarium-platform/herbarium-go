package haproxy

import (
	"fmt"
	"log"
)

// TransactionMiddleware is a middleware that manages transactions for HAProxy operations.
type TransactionMiddleware func(next func(transactionID string) error) func() error

// NewTransactionMiddleware creates a new TransactionMiddleware using the provided configManager interface.
func NewTransactionMiddleware(configManager HAProxyConfigurationManagerInterface) TransactionMiddleware {
	return func(next func(transactionID string) error) func() error {
		return func() error {
			// Retrieve the current config version using the interface method
			cfgVer, err := configManager.GetCurrentConfigVersion()
			if err != nil {
				log.Printf("[ERROR] Failed to get config version: %v", err)
				return fmt.Errorf("failed to retrieve configuration version: %v", err)
			}
			log.Printf("[INFO] Got config version: %d", cfgVer)

			// Start the transaction using the interface method
			transactionID, err := configManager.StartTransaction(cfgVer)
			if err != nil {
				log.Printf("[ERROR] Failed to start transaction: %v", err)
				return fmt.Errorf("failed to start transaction: %v", err)
			}
			log.Printf("[INFO] Started transaction: %s", transactionID)

			var executionErr error
			defer func() {
				// Rollback or commit the transaction depending on execution outcome
				if executionErr != nil {
					log.Printf("[ERROR] Rolling back transaction %s: %v", transactionID, executionErr)
					configManager.RollbackTransaction(transactionID)
				} else {
					log.Printf("[INFO] Committing transaction: %s", transactionID)
					configManager.CommitTransaction(transactionID)
				}
			}()

			log.Printf("[INFO] Executing operation with transaction: %s", transactionID)
			executionErr = next(transactionID)
			return executionErr
		}
	}
}
