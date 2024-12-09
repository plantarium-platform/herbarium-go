
# Herbarium

**Herbarium** is a core component of the [Plantarium Platform](https://github.com/plantarium-platform), a lightweight and resource-efficient solution inspired by cloud architecture principles, designed for running serverless functions and microservices.

Herbarium serves as the root of the platform, responsible for initializing system components, managing stems and leafs, and providing internal APIs for their management. It leverages an in-memory storage solution to optimize performance and enable real-time data handling.

## Key Features

- **System Initialization**: Bootstraps the platform and loads configurations.
- **Component Management**: Starts, stops, and monitors stems and leafs through internal APIs.
- **In-Memory Storage**: Provides fast, lightweight storage for managing platform state.
- **Configuration Management**: Reads and applies settings from a configuration file to control platform behavior.

## How It Works

### Initial Setup
- Herbarium initializes the platform by reading configurations from the `config.yaml` file.
- It sets up in-memory storage and prepares internal APIs for managing stems and leafs.
- The platform components (stems and leafs) are dynamically started based on the configurations.

### Internal APIs
- Herbarium exposes APIs to manage the lifecycle of stems and leafs:
    - **Start Stem/Leaf**: Dynamically start components based on demand.
    - **Stop Stem/Leaf**: Gracefully stop running components.
    - **Query State**: Retrieve information about the platform state and component statuses.

## Project Structure

Below is the simplified file structure of the Herbarium project:

```plaintext
.
├── cmd/herbarium
│   └── main.go                # Entry point for the Herbarium application
├── internal
│   ├── api/grpc               # Internal APIs for managing stems and leafs
│   ├── config                 # Configuration parsing and management
│   ├── haproxy                # HAProxy integration
│   ├── manager                # Logic for platform, stem, and leaf management
│   └── storage                # In-memory storage implementation
├── pkg
│   └── models                 # Shared models used across the project
├── testdata
│   ├── services               # Example service configurations and binaries
│   └── system/herbarium       # Example configuration for the Herbarium system
├── go.mod                     # Go module definition
├── go.sum                     # Go dependencies
└── README.md                  # Documentation (you are here)
```

## How to Run

1. **Step 1: Prepare Configuration**
    - Update the `testdata/system/herbarium/config.yaml` file with your deployment directory and platform settings.

2. **Step 2: Run Herbarium**
    - Start the application using the following command:
      ```bash
      go run cmd/herbarium/main.go
      ```

3. **Step 3: Verify Internal APIs**
    - Herbarium exposes gRPC APIs for managing stems and leafs. Use a gRPC client (e.g., `grpcurl`) to test the APIs:
      ```bash
      grpcurl -plaintext localhost:50051 <API_METHOD>
      ```

4. **Step 4: Build for Deployment**
    - To build a binary:
      ```bash
      go build -o herbarium cmd/herbarium/main.go
      ```
    - Run the binary:
      ```bash
      ./herbarium
      ```

## Testing

To run tests, use the following command:

```bash
go test ./...
```

## Contact

If you have questions or want to contribute, feel free to reach out on [GitHub](https://github.com/glorko) or [Telegram](https://t.me/glorfindeil).