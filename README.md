# Dante GPU Rental Platform

A decentralized GPU rental platform that enables users to rent GPU resources for AI training and computation using dGPU tokens on the Solana blockchain.

## Architecture Overview

The platform consists of multiple microservices working together to provide a complete GPU rental solution with real-time billing and blockchain integration.

### Core Services

1. **API Gateway**
   - Language: Go
   - Description: Central entry point handling authentication, routing, and billing endpoints. Provides unified access to all platform services with JWT authentication and rate limiting.
   - Status: Production Ready

2. **Auth Service**
   - Language: Python (FastAPI)
   - Description: JWT-based user authentication and authorization with role management. Handles user registration, login, and token management.
   - Status: Production Ready

3. **Provider Registry Service**
   - Language: Go
   - Description: GPU provider registration, discovery, and status management. Tracks provider hardware specifications, availability, and performance metrics.
   - Status: Production Ready

4. **Scheduler Orchestrator Service**
   - Language: Go
   - Description: Intelligent job scheduling with billing validation and provider selection. Integrates with billing service for cost validation before job execution.
   - Status: Production Ready

5. **Provider Daemon**
   - Language: Go
   - Description: Executes tasks on provider machines with real-time monitoring and billing integration. Supports Docker and script execution with GPU resource management.
   - Status: Production Ready

6. **Storage Service**
   - Language: Go
   - Description: S3-compatible file storage with presigned URL support. Handles user data, AI models, datasets, and job results using MinIO backend.
   - Status: Production Ready

7. **Billing Payment Service**
   - Language: Go
   - Description: Complete dGPU token payment system with Solana blockchain integration. Handles wallets, transactions, pricing, and real-time billing.
   - Status: Production Ready

8. **Monitoring Logging Service**
   - Language: Docker Compose Stack
   - Description: Comprehensive system monitoring with Prometheus, Grafana, and Loki. Provides real-time metrics, alerting, and log aggregation.
   - Status: Production Ready

### Key Features

#### Blockchain Integration
- Native dGPU token payments on Solana blockchain
- Real-time transaction verification and confirmation
- Automated wallet management and token transfers
- Platform fee collection and provider payouts

#### Dynamic Pricing Engine
- GPU model-specific base rates (RTX 4090, A100, H100, Apple Silicon, etc.)
- VRAM allocation-based pricing (per GB per hour)
- Power consumption multipliers
- Dynamic demand and supply adjustments
- Platform fee calculation (5% default)

#### Real-time Billing System
- Session-based billing with automatic monitoring
- Usage tracking with 1-minute intervals
- Insufficient funds protection with grace periods
- Automatic session termination on balance depletion
- Provider earnings calculation and distribution

#### GPU Marketplace
- Real-time GPU availability and pricing
- Advanced filtering by GPU type, VRAM, location
- Cost estimation before job submission
- Provider performance metrics and ratings

#### Secure Job Execution
- Docker and script-based execution environments
- GPU resource isolation and allocation
- Real-time performance monitoring
- Automatic cleanup and resource management

## Technology Stack

### Backend Services
- Go 1.21+ - Primary backend language for microservices
- Python 3.11+ - Auth service with FastAPI
- NATS JetStream - Message queue and event streaming
- Consul - Service discovery and configuration
- PostgreSQL - Primary database for all services

### Blockchain & Payments
- Solana - Blockchain platform for dGPU token
- SPL Token - dGPU token implementation (7xUV6YR3rZMfExPqZiovQSUxpnHxr2KJJqFg1bFrpump)
- Solana Go SDK - Blockchain integration library

### Storage & Monitoring
- MinIO - S3-compatible object storage
- Prometheus - Metrics collection and alerting
- Grafana - Monitoring dashboards and visualization
- Loki - Log aggregation and analysis

### Infrastructure
- Docker - Containerization and deployment
- Docker Compose - Local development orchestration

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Python 3.11 or higher
- Docker and Docker Compose
- PostgreSQL 14+
- NATS Server 2.9+
- Consul 1.15+
- MinIO (latest)

### Environment Setup

1. Clone the repository:
```bash
git clone https://github.com/dante-gpu/dante-backend.git
cd dante-backend
```

2. Set up Solana wallet for development:
```bash
export DEVELOPMENT_MODE=true
export SOLANA_PRIVATE_KEY="your_base58_private_key"
export DGPU_TOKEN_ADDRESS="7xUV6YR3rZMfExPqZiovQSUxpnHxr2KJJqFg1bFrpump"
```

3. Configure database and services:
```bash
export DATABASE_URL="postgres://dante:dante123@localhost:5432/dante"
export NATS_URL="nats://localhost:4222"
export CONSUL_URL="http://localhost:8500"
export JWT_SECRET="your_jwt_secret_key"
```

### Infrastructure Startup

1. Start core infrastructure:
```bash
docker-compose up -d infrastructure
```

2. Start monitoring stack:
```bash
cd monitoring-logging-service
docker-compose up -d
```

3. Initialize databases:
```bash
# Auth service migrations
cd auth-service && alembic upgrade head

# Other services will auto-migrate on startup
```

### Service Startup Order

Start services in the following order for proper dependency resolution:

1. Auth Service (port 8000)
2. Provider Registry Service (port 8001)
3. Billing Payment Service (port 8080)
4. Storage Service (port 8002)
5. Scheduler Orchestrator Service (port 8003)
6. API Gateway (port 8080)
7. Provider Daemon (on provider machines)

Refer to individual service READMEs for detailed setup instructions.

## Service Documentation

Each service includes comprehensive documentation:

- [API Gateway](./api-gateway/README.md) - Routing, authentication, billing endpoints
- [Auth Service](./auth-service/README.md) - User management and JWT authentication
- [Provider Registry Service](./provider-registry-service/README.md) - GPU provider management
- [Scheduler Orchestrator Service](./scheduler-orchestrator-service/README.md) - Job scheduling and billing integration
- [Provider Daemon](./provider-daemon/README.md) - Task execution and monitoring
- [Storage Service](./storage-service/README.md) - File storage and management
- [Billing Payment Service](./billing-payment-service/README.md) - dGPU token payments and blockchain
- [Monitoring Logging Service](./monitoring-logging-service/README.md) - System monitoring and alerting

## API Endpoints

### Authentication
- `POST /api/v1/auth/login` - User login with JWT token
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/refresh` - Token refresh

### Job Management
- `POST /api/v1/jobs` - Submit GPU rental job
- `GET /api/v1/jobs/{id}` - Get job status
- `DELETE /api/v1/jobs/{id}` - Cancel job

### Billing & Payments
- `POST /api/v1/billing/wallet` - Create dGPU wallet
- `GET /api/v1/billing/wallet/{id}/balance` - Check wallet balance
- `POST /api/v1/billing/wallet/{id}/deposit` - Deposit dGPU tokens
- `POST /api/v1/billing/wallet/{id}/withdraw` - Withdraw dGPU tokens
- `GET /api/v1/billing/marketplace` - Browse available GPUs
- `POST /api/v1/billing/pricing/estimate` - Estimate job cost

### Provider Management
- `POST /api/v1/providers` - Register GPU provider
- `GET /api/v1/providers` - List available providers
- `PUT /api/v1/providers/{id}/status` - Update provider status

### Storage
- `PUT /api/v1/storage/{bucket}/{key}` - Upload file
- `GET /api/v1/storage/{bucket}/{key}` - Download file
- `DELETE /api/v1/storage/{bucket}/{key}` - Delete file

## Configuration

### Environment Variables

Key environment variables for the platform:

```bash
# Solana Configuration
SOLANA_RPC_URL=https://api.mainnet-beta.solana.com
DGPU_TOKEN_ADDRESS=7xUV6YR3rZMfExPqZiovQSUxpnHxr2KJJqFg1bFrpump
SOLANA_PRIVATE_KEY=your_base58_private_key

# Database
DATABASE_URL=postgres://user:pass@localhost:5432/dante

# NATS
NATS_URL=nats://localhost:4222

# Consul
CONSUL_URL=http://localhost:8500

# JWT
JWT_SECRET=your_jwt_secret_key
```

### Pricing Configuration

Default GPU pricing rates (dGPU tokens per hour):

- NVIDIA RTX 4090: 0.50
- NVIDIA RTX 4080: 0.40
- NVIDIA A100: 2.00
- NVIDIA H100: 3.00
- Apple M3 Ultra: 1.20
- Apple M2 Ultra: 1.00

Additional rates:
- VRAM: 0.02 per GB per hour
- Power: 0.001 per watt per hour
- Platform fee: 5%

## Development

### Running Tests

```bash
# Run tests for all Go services
make test

# Run tests for specific service
cd billing-payment-service && go test ./...

# Run Python tests
cd auth-service && python -m pytest
```

### Building Services

```bash
# Build all services
make build

# Build specific service
cd api-gateway && go build -o bin/api-gateway cmd/main.go
```

### Database Migrations

```bash
# Run migrations for billing service
cd billing-payment-service && migrate -path migrations -database $DATABASE_URL up

# Run migrations for auth service
cd auth-service && alembic upgrade head
```

## Deployment

### Production Deployment

1. Configure production environment variables
2. Set up SSL certificates for HTTPS
3. Configure Solana mainnet endpoints
4. Set up database backups
5. Configure monitoring alerts
6. Deploy using Docker Compose or Kubernetes

### Security Considerations

- Use strong JWT secrets in production
- Secure Solana private keys with hardware wallets
- Enable database encryption at rest
- Configure network firewalls
- Regular security audits and updates

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go and Python coding standards
- Add comprehensive tests for new features
- Update documentation for API changes
- Ensure backward compatibility
- Test blockchain integration thoroughly

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support and questions:
- Create an issue on GitHub
- Join our Discord community
- Check the documentation wiki
