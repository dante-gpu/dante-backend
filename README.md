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

## How the Platform Works

### User Journey Overview

The Dante GPU Rental Platform operates as a decentralized marketplace connecting GPU providers with users who need computational resources for AI training and other GPU-intensive tasks.

#### For GPU Providers

1. **Registration and Setup**
   ```bash
   # Provider installs the daemon on their machine
   ./provider-daemon --register
   ```
   - Provider registers their GPU hardware (model, VRAM, location)
   - Sets custom pricing rates or uses platform defaults
   - Configures availability schedule and resource limits

2. **Resource Monitoring**
   - Provider daemon continuously monitors GPU status and availability
   - Reports real-time metrics to the platform (temperature, utilization, power)
   - Automatically updates availability based on local usage

3. **Job Execution**
   - Receives job assignments through NATS message queue
   - Executes tasks in isolated Docker containers or script environments
   - Monitors resource usage and reports billing data in real-time

4. **Earnings and Payouts**
   - Earns dGPU tokens based on actual usage time and resource allocation
   - Platform automatically calculates earnings (base rate + VRAM + power consumption)
   - Receives automatic payouts to Solana wallet after job completion

#### For GPU Renters

1. **Account Setup and Wallet**
   ```bash
   # User creates account and dGPU wallet
   curl -X POST /api/v1/auth/register -d '{"email":"user@example.com","password":"secure123"}'
   curl -X POST /api/v1/billing/wallet -H "Authorization: Bearer $JWT_TOKEN"
   ```

2. **Browse GPU Marketplace**
   ```bash
   # Browse available GPUs with filtering
   curl "/api/v1/billing/marketplace?gpu_model=RTX4090&min_vram=16&location=US"
   ```
   - Filter by GPU model, VRAM, location, and price range
   - View real-time availability and current pricing
   - Check provider ratings and performance history

3. **Cost Estimation**
   ```bash
   # Estimate job cost before submission
   curl -X POST /api/v1/billing/pricing/estimate -d '{
     "gpu_model": "RTX4090",
     "vram_gb": 12,
     "estimated_hours": 2.5,
     "power_watts": 350
   }'
   ```

4. **Job Submission and Payment**
   ```bash
   # Submit GPU rental job
   curl -X POST /api/v1/jobs -d '{
     "name": "AI Model Training",
     "docker_image": "pytorch/pytorch:latest",
     "script": "python train.py",
     "gpu_requirements": {
       "model": "RTX4090",
       "vram_gb": 12,
       "min_compute_capability": 8.6
     },
     "max_duration_hours": 4,
     "max_cost_dgpu": 2.5
   }'
   ```

5. **Real-time Monitoring**
   - Monitor job progress and resource usage through API
   - Receive real-time billing updates
   - Get notifications for job completion or issues

### Platform Workflow

#### Job Lifecycle

1. **Job Submission**
   ```
   User → API Gateway → Scheduler Service → Job Queue (NATS)
   ```

2. **Provider Matching**
   ```
   Scheduler → Provider Registry → Billing Validation → Provider Selection
   ```

3. **Job Execution**
   ```
   Scheduler → Provider Daemon → Docker/Script Execution → Real-time Monitoring
   ```

4. **Billing and Completion**
   ```
   Provider Daemon → Billing Service → Solana Blockchain → Payment Processing
   ```

#### Real-time Billing Process

1. **Session Initialization**
   - Billing service creates a rental session when job starts
   - Validates user has sufficient dGPU token balance
   - Sets up real-time usage monitoring

2. **Usage Tracking**
   ```
   Provider Daemon → NATS → Billing Service (every minute)
   ```
   - Tracks actual GPU usage, VRAM allocation, and power consumption
   - Calculates incremental costs based on dynamic pricing
   - Monitors user balance and enforces spending limits

3. **Payment Processing**
   - Automatically deducts dGPU tokens from user wallet
   - Transfers provider earnings to escrow
   - Handles platform fee collection (5% default)

4. **Session Completion**
   - Finalizes billing calculations
   - Processes provider payout to Solana wallet
   - Generates usage reports and transaction records

### Dynamic Pricing Algorithm

#### Base Pricing Structure
```
Total Cost = (Base Rate + VRAM Cost + Power Cost) × Time × Demand Multiplier
```

#### Pricing Components

1. **GPU Base Rates** (dGPU tokens per hour)
   ```
   RTX 4090: 0.50    A100: 2.00    H100: 3.00
   RTX 4080: 0.40    A6000: 1.50   Apple M3 Ultra: 1.20
   ```

2. **VRAM Allocation** (per GB per hour)
   ```
   VRAM Cost = Allocated VRAM (GB) × 0.02 dGPU tokens
   ```

3. **Power Consumption** (per watt per hour)
   ```
   Power Cost = GPU Power Draw (watts) × 0.001 dGPU tokens
   ```

4. **Dynamic Demand Multiplier**
   ```
   High Demand (>80% utilization): 1.2x
   Normal Demand (20-80%): 1.0x
   Low Demand (<20%): 0.8x
   ```

### Blockchain Integration

#### dGPU Token Operations

1. **Wallet Creation**
   ```solana
   // Automatically creates SPL token account for user
   Token Address: 7xUV6YR3rZMfExPqZiovQSUxpnHxr2KJJqFg1bFrpump
   ```

2. **Payment Processing**
   ```
   User Wallet → Platform Escrow → Provider Wallet
   ```

3. **Transaction Verification**
   - All payments verified on Solana blockchain
   - Real-time transaction confirmation
   - Automatic retry for failed transactions

#### Security Features

- Multi-signature wallets for large transactions
- Escrow system for job payments
- Automatic refunds for failed jobs
- Fraud detection and prevention

### Monitoring and Observability

#### Real-time Metrics
- GPU utilization and temperature
- Job execution progress
- Billing and payment status
- Provider performance metrics

#### Alerting System
- Job failure notifications
- Payment processing alerts
- Provider availability changes
- System health monitoring

### API Integration Examples

#### Complete Job Submission Flow
```bash
# 1. Authenticate user
JWT_TOKEN=$(curl -X POST /api/v1/auth/login -d '{"email":"user@example.com","password":"secure123"}' | jq -r '.access_token')

# 2. Check wallet balance
curl -H "Authorization: Bearer $JWT_TOKEN" /api/v1/billing/user/123/balance

# 3. Browse available GPUs
curl -H "Authorization: Bearer $JWT_TOKEN" "/api/v1/billing/marketplace?gpu_model=RTX4090"

# 4. Estimate job cost
COST_ESTIMATE=$(curl -X POST -H "Authorization: Bearer $JWT_TOKEN" /api/v1/billing/pricing/estimate -d '{
  "gpu_model": "RTX4090",
  "vram_gb": 12,
  "estimated_hours": 2
}')

# 5. Submit job
JOB_ID=$(curl -X POST -H "Authorization: Bearer $JWT_TOKEN" /api/v1/jobs -d '{
  "name": "AI Training Job",
  "docker_image": "pytorch/pytorch:latest",
  "script": "python train.py",
  "gpu_requirements": {"model": "RTX4090", "vram_gb": 12},
  "max_cost_dgpu": 1.5
}' | jq -r '.job_id')

# 6. Monitor job progress
curl -H "Authorization: Bearer $JWT_TOKEN" /api/v1/jobs/$JOB_ID

# 7. Check real-time billing
curl -H "Authorization: Bearer $JWT_TOKEN" /api/v1/billing/sessions/$JOB_ID/usage
```

This comprehensive workflow demonstrates how the Dante GPU Rental Platform creates a seamless, secure, and efficient marketplace for GPU resources, powered by blockchain technology and real-time billing.

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
