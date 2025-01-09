package haproxy

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"strconv"
)

// HAProxyServer struct represents a backend server in HAProxy.
type HAProxyServer struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Port    int    `json:"port"`
}

// HAProxyConfigurationManagerInterface defines the methods for managing HAProxy configuration.
type HAProxyConfigurationManagerInterface interface {
	GetCurrentConfigVersion() (int64, error)
	StartTransaction(version int64) (string, error)
	CommitTransaction(transactionID string) error
	RollbackTransaction(transactionID string) error
	CreateBackend(backendName, transactionID string) error
	AddServer(backendName, serverName, host string, port int, transactionID string) error
	DeleteServer(backendName, serverName, transactionID string) error
	GetServersFromBackend(backendName, transactionID string) ([]HAProxyServer, error)
}

// HAProxyConfigurationManager is the concrete implementation of HAProxyConfigurationManagerInterface.
type HAProxyConfigurationManager struct {
	client *resty.Client
}

// NewHAProxyConfigurationManager initializes the configuration manager with the provided HAProxyConfig.
func NewHAProxyConfigurationManager(config HAProxyConfig) *HAProxyConfigurationManager {
	client := resty.New()
	client.SetBaseURL(config.APIURL)
	client.SetBasicAuth(config.Username, config.Password)
	client.SetHeader("Content-Type", "application/json")
	client.SetDisableWarn(true)

	return &HAProxyConfigurationManager{
		client: client,
	}
}

// GetCurrentConfigVersion retrieves the current HAProxy configuration version as an integer.
func (c *HAProxyConfigurationManager) GetCurrentConfigVersion() (int64, error) {
	resp, err := c.client.R().Get("/configuration/version")
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve version: %v", err)
	}

	if resp.StatusCode() != 200 {
		return 0, fmt.Errorf("failed to retrieve version, status code: %d, response: %s", resp.StatusCode(), resp.String())
	}

	version, err := strconv.ParseInt(resp.String(), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse version as integer: %v", err)
	}

	return version, nil
}

// StartTransaction starts a new HAProxy configuration transaction.
func (c *HAProxyConfigurationManager) StartTransaction(version int64) (string, error) {
	resp, err := c.client.R().SetQueryParam("version", strconv.FormatInt(version, 10)).Post("/transactions")
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %v", err)
	}

	if resp.StatusCode() != 201 {
		return "", fmt.Errorf("failed to start transaction, status code: %d, response: %s", resp.StatusCode(), resp.String())
	}

	var transaction struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(resp.Body(), &transaction); err != nil {
		return "", fmt.Errorf("failed to parse transaction ID: %v", err)
	}

	return transaction.ID, nil
}

// CommitTransaction commits the specified HAProxy configuration transaction.
func (c *HAProxyConfigurationManager) CommitTransaction(transactionID string) error {
	resp, err := c.client.R().Put(fmt.Sprintf("/transactions/%s", transactionID))
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	if resp.StatusCode() != 202 {
		return fmt.Errorf("failed to commit transaction, status code: %d, response: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

// RollbackTransaction rolls back the specified HAProxy configuration transaction.
func (c *HAProxyConfigurationManager) RollbackTransaction(transactionID string) error {
	resp, err := c.client.R().Delete(fmt.Sprintf("/transactions/%s", transactionID))
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %v", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to rollback transaction, status code: %d, response: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

// CreateBackend creates a new backend in the HAProxy configuration.
func (c *HAProxyConfigurationManager) CreateBackend(backendName, transactionID string) error {
	log.Printf("[HAProxyConfigurationManager] Checking if backend exists: backendName=%s, transactionID=%s", backendName, transactionID)

	// Check if the backend exists by name
	resp, err := c.client.R().
		SetQueryParam("transaction_id", transactionID).
		Get(fmt.Sprintf("/configuration/backends/%s", backendName))
	if err != nil {
		log.Printf("[HAProxyConfigurationManager] Error checking backend existence: backendName=%s, transactionID=%s, error=%v", backendName, transactionID, err)
		return fmt.Errorf("failed to check if backend exists: %v", err)
	}

	log.Printf("[HAProxyConfigurationManager] Backend existence check response: statusCode=%d, responseBody=%s", resp.StatusCode(), resp.String())

	// If the backend exists, delete it
	if resp.StatusCode() == 200 {
		log.Printf("[HAProxyConfigurationManager] Backend %s exists. Deleting backend...", backendName)

		deleteResp, err := c.client.R().
			SetQueryParam("transaction_id", transactionID).
			Delete(fmt.Sprintf("/configuration/backends/%s", backendName))
		if err != nil {
			log.Printf("[HAProxyConfigurationManager] Error deleting backend: backendName=%s, transactionID=%s, error=%v", backendName, transactionID, err)
			return fmt.Errorf("failed to delete existing backend: %v", err)
		}

		log.Printf("[HAProxyConfigurationManager] Backend deletion response: statusCode=%d, responseBody=%s", deleteResp.StatusCode(), deleteResp.String())

		if deleteResp.StatusCode() != 202 {
			log.Printf("[HAProxyConfigurationManager] Unexpected status code while deleting backend: backendName=%s, transactionID=%s, statusCode=%d, responseBody=%s",
				backendName, transactionID, deleteResp.StatusCode(), deleteResp.String())
			return fmt.Errorf("unexpected status code while deleting backend: %d, response: %s", deleteResp.StatusCode(), deleteResp.String())
		}
		log.Printf("[HAProxyConfigurationManager] Backend %s deleted successfully.", backendName)
	}

	// Create a new backend
	log.Printf("[HAProxyConfigurationManager] Creating backend: %s", backendName)

	backendData := map[string]interface{}{
		"name": backendName,
		"mode": "http",
		"balance": map[string]string{
			"algorithm": "roundrobin",
		},
		"http_connection_mode": "http-server-close",
		"redispatch": map[string]interface{}{
			"enabled": "enabled",
		},
		"http-check": map[string]interface{}{
			"method":  "HEAD",
			"uri":     "/",
			"version": "HTTP/1.1",
			"headers": []map[string]string{
				{
					"name":  "Host",
					"value": "localhost",
				},
			},
		},
	}

	createResp, err := c.client.R().
		SetQueryParam("transaction_id", transactionID).
		SetBody(backendData).
		Post("/configuration/backends")
	if err != nil {
		log.Printf("[HAProxyConfigurationManager] Error creating backend: backendName=%s, transactionID=%s, error=%v", backendName, transactionID, err)
		return fmt.Errorf("failed to create backend: %v", err)
	}

	log.Printf("[HAProxyConfigurationManager] Backend creation response: statusCode=%d, responseBody=%s", createResp.StatusCode(), createResp.String())

	if createResp.StatusCode() != 202 {
		log.Printf("[HAProxyConfigurationManager] Unexpected status code while creating backend: backendName=%s, transactionID=%s, statusCode=%d, responseBody=%s",
			backendName, transactionID, createResp.StatusCode(), createResp.String())
		return fmt.Errorf("unexpected status code while creating backend: %d, response: %s", createResp.StatusCode(), createResp.String())
	}

	log.Printf("[HAProxyConfigurationManager] Backend %s created successfully", backendName)
	return nil
}

// AddServer adds a new server to the specified backend in the HAProxy configuration.
func (c *HAProxyConfigurationManager) AddServer(backendName, serverName, host string, port int, transactionID string) error {
	resp, err := c.client.R().
		SetQueryParam("transaction_id", transactionID).
		SetBody(map[string]interface{}{
			"name":    serverName,
			"address": host,
			"port":    port,
		}).
		Post(fmt.Sprintf("/configuration/backends/%s/servers", backendName))
	if err != nil {
		return fmt.Errorf("failed to add server to backend %s: %v", backendName, err)
	}

	// Analyze the response
	if resp.StatusCode() != 202 && resp.StatusCode() != 201 {
		return fmt.Errorf(
			"unexpected status code %d when adding server to backend %s: response: %s",
			resp.StatusCode(), backendName, resp.String(),
		)
	}

	log.Printf("[HAProxyConfigurationManager] Server %s (host=%s, port=%d) added to backend %s successfully. Status: %d", serverName, host, port, backendName, resp.StatusCode())
	return nil
}

// DeleteServer deletes a specific server from the backend.
func (c *HAProxyConfigurationManager) DeleteServer(backendName, serverName, transactionID string) error {
	resp, err := c.client.R().
		SetQueryParam("transaction_id", transactionID).
		Delete(fmt.Sprintf("/configuration/backends/%s/servers/%s", backendName, serverName))

	if err != nil {
		return fmt.Errorf("failed to delete server %s from backend %s: %v", serverName, backendName, err)
	}

	switch resp.StatusCode() {
	case 204, 202: // Accept both immediate success and accepted for reload
		log.Printf("[INFO] Server %s successfully deleted from backend %s", serverName, backendName)
		return nil
	case 404:
		var apiErr struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(resp.Body(), &apiErr); err != nil {
			return fmt.Errorf("failed to parse error response: %v", err)
		}
		log.Printf("[INFO] Server or backend not found: %s", apiErr.Message)
		return nil
	case 400:
		return fmt.Errorf("API error deleting server %s from backend %s: %s", serverName, backendName, resp.String())
	default:
		return fmt.Errorf("unexpected status %d deleting server %s from backend %s: %s", resp.StatusCode(), serverName, backendName, resp.String())
	}
}

// GetServersFromBackend retrieves all servers from a specified backend in the HAProxy configuration.
func (c *HAProxyConfigurationManager) GetServersFromBackend(backendName, transactionID string) ([]HAProxyServer, error) {
	resp, err := c.client.R().
		SetQueryParam("transaction_id", transactionID).
		Get(fmt.Sprintf("/configuration/backends/%s/servers", backendName))
	if err != nil {
		return nil, fmt.Errorf("failed to list servers in backend %s: %v", backendName, err)
	}

	if resp.StatusCode() == 404 {
		log.Printf("[INFO] Backend %s not found, no servers to get", backendName)
		return nil, nil
	} else if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to list servers, status code: %d, response: %s", resp.StatusCode(), resp.String())
	}

	var servers []HAProxyServer
	if err := json.Unmarshal(resp.Body(), &servers); err != nil {
		return nil, fmt.Errorf("failed to parse server list: %v", err)
	}

	return servers, nil
}
