# Dante GPU Rental Platform - Development Guide

## Overview

This document provides guidance for developing and testing the Dante GPU Rental Platform using the three main executables.

## Project Structure

```
dante-backend/
├── common/                    # Shared types and utilities
│   ├── types.go              # Common data structures
│   └── utils.go              # Shared utility functions
├── cmd/                      # Main executables
│   ├── provider/main.go      # GPU Provider daemon
│   ├── rental/main.go        # GPU Rental client
│   └── test-server/main.go   # Development test server
├── (microservices)...        # Production microservices
└── go.mod                    # Go module definition
```

## Applications

### 1. GPU Provider Daemon (`cmd/provider/main.go`)

A complete GPU provider implementation that:
- Automatically detects available GPUs (NVIDIA, Apple Silicon, AMD)
- Registers with the provider registry service
- Listens for job assignments via NATS
- Executes tasks using Docker or direct script execution
- Reports real-time billing and usage metrics
- Handles graceful shutdown

**Building and Running:**
```bash
# Build
go build -o provider-daemon ./cmd/provider

# Run with environment variables
export PROVIDER_REGISTRY_URL="http://localhost:8001"
export BILLING_SERVICE_URL="http://localhost:8003"
export NATS_ADDRESS="nats://localhost:4222"
./provider-daemon

# Or run directly
go run ./cmd/provider
```

**Key Features:**
- Multi-platform GPU detection
- Real-time metrics reporting
- Job execution with Docker support
- Graceful shutdown handling
- NATS-based task communication

### 2. GPU Rental Client (`cmd/rental/main.go`)

A comprehensive rental client that provides:
- JWT authentication with the API Gateway
- Wallet management and balance checking
- Provider discovery and filtering
- Job submission (AI training, custom scripts)
- Cost estimation and pricing
- Job monitoring and cancellation
- Interactive menu system and command-line interface

**Building and Running:**
```bash
# Build
go build -o gpu-rental-client ./cmd/rental

# Interactive mode
./gpu-rental-client

# Command line mode
./gpu-rental-client providers          # List providers
./gpu-rental-client balance           # Check balance
./gpu-rental-client submit "my-job"   # Submit job
./gpu-rental-client status <job-id>   # Check status

# With environment variables
export API_GATEWAY_URL="http://localhost:8080"
export DANTE_USERNAME="your-username"
export DANTE_PASSWORD="your-password"
./gpu-rental-client
```

**Key Features:**
- Interactive menu-driven interface
- Command-line automation support
- Complete job lifecycle management
- Real-time cost estimation
- Wallet and billing integration

### 3. Test Server (`cmd/test-server/main.go`)

A development server for testing and demonstration:
- Mock service endpoints
- Sample GPU provider data
- Health checks and status monitoring
- Demo registration and provider listing

**Building and Running:**
```bash
# Build
go build -o test-server ./cmd/test-server

# Run
./test-server
# Server starts on http://localhost:9999

# Or run directly
go run ./cmd/test-server
```

**Available Endpoints:**
- `GET /health` - Health check
- `GET /services` - Service status
- `GET /demo` - Demo platform information
- `POST /provider/register` - Register demo provider
- `GET /providers` - List demo providers
- `POST /provider/status` - Update provider status

## Development Workflow

### 1. Local Development Setup

```bash
# Install dependencies
go mod tidy

# Start the test server
go run ./cmd/test-server

# In another terminal, start the provider daemon
go run ./cmd/provider

# In a third terminal, run the rental client
go run ./cmd/rental
```

### 2. Testing Provider Registration

```bash
# Test provider registration
curl -X POST http://localhost:9999/provider/register

# Check registered providers
curl http://localhost:9999/providers
```

### 3. Environment Configuration

Create a `.env` file for local development:
```bash
# Service URLs
API_GATEWAY_URL=http://localhost:8080
PROVIDER_REGISTRY_URL=http://localhost:8001
BILLING_SERVICE_URL=http://localhost:8003
STORAGE_SERVICE_URL=http://localhost:8082
NATS_ADDRESS=nats://localhost:4222

# Credentials
DANTE_USERNAME=demo-user
DANTE_PASSWORD=demo-pass

# Provider Settings
PROVIDER_NAME=Local-Dev-Provider
SOLANA_WALLET_ADDRESS=11111111111111111111111111111111
```

### 4. Build All Executables

```bash
# Build all executables
make build-all
# Or manually:
go build -o provider-daemon ./cmd/provider
go build -o gpu-rental-client ./cmd/rental
go build -o test-server ./cmd/test-server
```

## Integration with Microservices

These applications are designed to integrate with the full Dante platform:

1. **API Gateway** (`api-gateway/`) - Central routing and authentication
2. **Auth Service** (`auth-service/`) - User authentication and JWT tokens
3. **Provider Registry** (`provider-registry-service/`) - Provider management
4. **Scheduler Orchestrator** (`scheduler-orchestrator-service/`) - Job scheduling
5. **Billing Payment** (`billing-payment-service/`) - dGPU token billing
6. **Storage Service** (`storage-service/`) - File and result storage
7. **Monitoring/Logging** (`monitoring-logging-service/`) - System monitoring

## Production Deployment

For production deployment:

1. Use proper service discovery (Consul)
2. Configure NATS clustering
3. Set up PostgreSQL databases
4. Deploy with Docker/Kubernetes
5. Configure monitoring and alerting
6. Set up proper Solana wallet integration

## GPU Detection

The provider daemon automatically detects:

- **NVIDIA GPUs** via `nvidia-smi`
- **Apple Silicon** via `system_profiler`
- **AMD GPUs** via `rocm-smi` (basic support)

## Troubleshooting

### Common Issues

1. **NATS Connection Failed**
   - Ensure NATS server is running on `localhost:4222`
   - Check network connectivity

2. **GPU Detection Failed**
   - Install proper GPU drivers
   - Verify GPU tools are in PATH

3. **Authentication Failed**
   - Check username/password
   - Verify API Gateway is running

4. **Build Errors**
   - Run `go mod tidy`
   - Check Go version (requires Go 1.21+)

### Debugging

Enable debug logging by setting:
```bash
export LOG_LEVEL=debug
```

## Next Steps

1. Integrate with Docker for containerized deployment
2. Add comprehensive test coverage
3. Implement advanced GPU scheduling algorithms
4. Add support for multi-GPU jobs
5. Enhance monitoring and alerting
6. Implement proper Solana blockchain integration 