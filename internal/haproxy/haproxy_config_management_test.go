package haproxy

import (
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestGetCurrentConfigVersion tests the GetCurrentConfigVersion method
func TestGetCurrentConfigVersion(t *testing.T) {
	// Initialize resty client
	client := resty.New()

	// Activate httpmock for the resty client's HTTP client
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Register a mock responder for the GET request
	httpmock.RegisterResponder("GET", "/configuration/version",
		httpmock.NewStringResponder(200, "1"))

	// Initialize the manager with the mocked client
	manager := &HAProxyConfigurationManager{
		client: client,
	}

	// Run the method under test
	version, err := manager.GetCurrentConfigVersion()

	// Assert the result
	assert.NoError(t, err)
	assert.Equal(t, int64(1), version)
}

func TestStartTransaction(t *testing.T) {
	// Initialize resty client
	client := resty.New()

	// Activate httpmock for the resty client's HTTP client
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Register a mock responder for the POST request
	httpmock.RegisterResponder("POST", "/transactions",
		httpmock.NewStringResponder(201, `{"id":"txn123"}`))

	// Initialize the manager with the mocked client
	manager := &HAProxyConfigurationManager{
		client: client,
	}

	// Run the method under test
	transactionID, err := manager.StartTransaction(1)

	// Assert the result
	assert.NoError(t, err)
	assert.Equal(t, "txn123", transactionID)
}

func TestCommitTransaction(t *testing.T) {
	// Initialize resty client
	client := resty.New()

	// Activate httpmock for the resty client's HTTP client
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Register a mock responder for the PUT request
	httpmock.RegisterResponder("PUT", "/transactions/txn123",
		httpmock.NewStringResponder(202, "{}"))

	// Initialize the manager with the mocked client
	manager := &HAProxyConfigurationManager{
		client: client,
	}

	// Run the method under test
	err := manager.CommitTransaction("txn123")

	// Assert the result
	assert.NoError(t, err)
}

func TestRollbackTransaction(t *testing.T) {
	// Initialize resty client
	client := resty.New()

	// Activate httpmock for the resty client's HTTP client
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Register a mock responder for the DELETE request
	httpmock.RegisterResponder("DELETE", "/transactions/txn123",
		httpmock.NewStringResponder(200, "{}"))

	// Initialize the manager with the mocked client
	manager := &HAProxyConfigurationManager{
		client: client,
	}

	// Run the method under test
	err := manager.RollbackTransaction("txn123")

	// Assert the result
	assert.NoError(t, err)
}

func TestCreateBackend(t *testing.T) {
	// Initialize resty client
	client := resty.New()

	// Activate httpmock for the resty client's HTTP client
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Register a mock responder for the GET request to check backend existence
	httpmock.RegisterResponder("GET", "/configuration/backends",
		httpmock.NewStringResponder(404, "")) // Simulate backend not found

	// Register a mock responder for the POST request to create a backend
	httpmock.RegisterResponder("POST", "/configuration/backends",
		httpmock.NewStringResponder(200, "{}"))

	// Initialize the manager with the mocked client
	manager := &HAProxyConfigurationManager{
		client: client,
	}

	// Run the method under test
	err := manager.CreateBackend("backend1", "txn123")

	// Assert the result
	assert.NoError(t, err)
}

func TestAddServer(t *testing.T) {
	// Initialize resty client
	client := resty.New()

	// Activate httpmock for the resty client's HTTP client
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Register a mock responder for the POST request to add a server
	httpmock.RegisterResponder("POST", "/configuration/backends/backend1/servers",
		httpmock.NewStringResponder(200, "{}"))

	// Initialize the manager with the mocked client
	manager := &HAProxyConfigurationManager{
		client: client,
	}

	// Run the method under test
	err := manager.AddServer("backend1", "server1", "localhost", "txn123")

	// Assert the result
	assert.NoError(t, err)
}

func TestDeleteServer(t *testing.T) {
	// Initialize resty client
	client := resty.New()

	// Activate httpmock for the resty client's HTTP client
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Register a mock responder for the DELETE request to delete a server
	httpmock.RegisterResponder("DELETE", "/configuration/backends/backend1/servers/server1",
		httpmock.NewStringResponder(204, "{}"))

	// Initialize the manager with the mocked client
	manager := &HAProxyConfigurationManager{
		client: client,
	}

	// Run the method under test
	err := manager.DeleteServer("backend1", "server1", "txn123")

	// Assert the result
	assert.NoError(t, err)
}

func TestGetServersFromBackend(t *testing.T) {
	// Initialize resty client
	client := resty.New()

	// Activate httpmock for the resty client's HTTP client
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	// Register a mock responder for the GET request to fetch servers from a backend
	httpmock.RegisterResponder("GET", "/configuration/backends/backend1/servers",
		httpmock.NewStringResponder(200, `[{"name":"server1","address":"localhost","port":8080}]`))

	// Initialize the manager with the mocked client
	manager := &HAProxyConfigurationManager{
		client: client,
	}

	// Run the method under test
	servers, err := manager.GetServersFromBackend("backend1", "txn123")

	// Assert the result
	assert.NoError(t, err)
	assert.Len(t, servers, 1)
	assert.Equal(t, "server1", servers[0].Name)
}
