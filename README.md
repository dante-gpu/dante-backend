# Dante GPU Rental Platform

A decentralized GPU rental platform that enables users to rent GPU resources for AI training and computation using dGPU tokens on the Solana blockchain.

## Architecture Overview

The platform consists of multiple microservices working together to provide a complete GPU rental solution with real-time billing and blockchain integration.

### Core Services

1. **API Gateway**
   - Language: Go
   - Description: Central entry point handling authentication, routing, and billing endpoints. Provides unified access to all platform services with JWT authentication and rate limiting.
   - Status: Production Ready ✅

2. **Auth Service**
   - Language: Python (FastAPI)
   - Description: JWT-based user authentication and authorization with role management. Handles user registration, login, and token management.
   - Status: Production Ready ✅

3. **Provider Registry Service**
   - Language: Go
   - Description: GPU provider registration, discovery, and status management. Tracks provider hardware specifications, availability, and performance metrics.
   - Status: Production Ready ✅

4. **Scheduler Orchestrator Service**
   - Language: Go
   - Description: Intelligent job scheduling with billing validation and provider selection. Integrates with billing service for cost validation before job execution.
   - Status: Production Ready ✅

5. **Provider Daemon**
   - Language: Go
   - Description: Executes tasks on provider machines with real-time monitoring and billing integration. Supports Docker and script execution with GPU resource management.
   - Status: Production Ready ✅

6. **Storage Service**
   - Language: Go
   - Description: S3-compatible file storage with presigned URL support. Handles user data, AI models, datasets, and job results using MinIO backend.
   - Status: Production Ready ✅

7. **Billing Payment Service**
   - Language: Go
   - Description: Complete dGPU token payment system with Solana blockchain integration. Handles wallets, transactions, pricing, and real-time billing.
   - Status: Production Ready ✅

8. **Frontend Web Application**
   - Language: Next.js (TypeScript/React)
   - Description: Modern web interface for users and providers. Features real-time dashboard, GPU marketplace, job management, and wallet integration.
   - Status: Production Ready ✅

9. **Monitoring Logging Service**
   - Language: Docker Compose Stack
   - Description: Comprehensive system monitoring with Prometheus, Grafana, and Loki. Provides real-time metrics, alerting, and log aggregation.
   - Status: Production Ready ✅

## 🚀 Quick Start - Production Deployment

### Prerequisites

- Docker 20.10+ and Docker Compose 2.0+
- 4GB+ RAM and 20GB+ storage
- Solana wallet with dGPU tokens for testing

### 1. Clone and Setup

```bash
git clone https://github.com/dante-gpu/dante-backend.git
cd dante-backend

# Copy environment configuration
cp env.production.example .env

# Edit with your production values
nano .env
```

### 2. Configure Environment

Update `.env` file with your production settings:

```bash
# Required: Set secure passwords
POSTGRES_PASSWORD=your_secure_postgres_password
JWT_SECRET=your_super_secure_jwt_secret_key_256_bits_minimum
MINIO_ROOT_PASSWORD=your_secure_minio_password

# Required: Solana configuration
SOLANA_PRIVATE_KEY=your_base58_encoded_solana_private_key
DGPU_TOKEN_ADDRESS=7xUV6YR3rZMfExPqZiovQSUxpnHxr2KJJqFg1bFrpump

# Optional: Production URLs
NEXT_PUBLIC_API_URL=https://api.yourdomain.com
```

### 3. Deploy Platform

```bash
# Make deployment script executable
chmod +x deploy-production.sh

# Deploy complete platform
./deploy-production.sh
```

The deployment script will:
- ✅ Check prerequisites
- ✅ Build all Docker images
- ✅ Deploy infrastructure services (Postgres, NATS, Consul, MinIO, Redis)
- ✅ Deploy application services (Auth, Billing, Provider Registry, Storage, Scheduler)
- ✅ Deploy API Gateway and Frontend
- ✅ Deploy monitoring stack (Prometheus, Grafana, Loki)
- ✅ Run health checks and validation tests

### 4. Access Services

After successful deployment:

| Service | URL | Credentials |
|---------|-----|-------------|
| **Frontend Dashboard** | http://localhost:3000 | demo/demo123 |
| **API Gateway** | http://localhost:8080 | - |
| **Grafana Monitoring** | http://localhost:3001 | admin/admin |
| **Prometheus Metrics** | http://localhost:9090 | - |
| **Consul Service Discovery** | http://localhost:8500 | - |
| **MinIO Storage Console** | http://localhost:9001 | dante/dante123456 |

### 5. Test Platform

```bash
# Test API Gateway health
curl http://localhost:8080/health

# Test Frontend health  
curl http://localhost:3000/api/health

# Test user authentication
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"demo","password":"demo123"}'

# Check GPU marketplace
curl http://localhost:8080/api/v1/billing/marketplace
```

## Key Features

### ✅ Blockchain Integration
- Native dGPU token payments on Solana blockchain
- Real-time transaction verification and confirmation
- Automated wallet management and token transfers
- Platform fee collection and provider payouts

### ✅ Dynamic Pricing Engine
- GPU model-specific base rates (RTX 4090, A100, H100, Apple Silicon, etc.)
- VRAM allocation-based pricing (per GB per hour)
- Power consumption multipliers
- Dynamic demand and supply adjustments
- Platform fee calculation (5% default)

### ✅ Real-time Billing System
- Session-based billing with automatic monitoring
- Usage tracking with 1-minute intervals
- Insufficient funds protection with grace periods
- Automatic session termination on balance depletion
- Provider earnings calculation and distribution

### ✅ GPU Marketplace
- Real-time GPU availability and pricing
- Advanced filtering by GPU type, VRAM, location
- Cost estimation before job submission
- Provider performance metrics and ratings

### ✅ Secure Job Execution
- Docker and script-based execution environments
- GPU resource isolation and allocation
- Real-time performance monitoring
- Automatic cleanup and resource management

### ✅ Modern Web Interface
- React/Next.js frontend with TypeScript
- Real-time dashboard with live updates
- Responsive design for desktop and mobile
- Comprehensive provider and user management

### ✅ Production-Ready Infrastructure
- Microservices architecture with Docker
- Service discovery with Consul
- Message queuing with NATS JetStream
- Monitoring with Prometheus and Grafana
- Logging with Loki
- Load balancing and health checks

## Technology Stack

### Backend Services
- **Go 1.21+** - Primary backend language for microservices
- **Python 3.11+** - Auth service with FastAPI
- **NATS JetStream** - Message queue and event streaming
- **Consul** - Service discovery and configuration
- **PostgreSQL** - Primary database for all services

### Frontend
- **Next.js 13+** - React framework with TypeScript
- **Tailwind CSS** - Utility-first CSS framework
- **shadcn/ui** - Modern component library
- **React Query** - Data fetching and caching

### Blockchain & Payments
- **Solana** - Blockchain platform for dGPU token
- **SPL Token** - dGPU token implementation
- **Solana Go SDK** - Blockchain integration library

### Storage & Monitoring
- **MinIO** - S3-compatible object storage
- **Redis** - Caching and rate limiting
- **Prometheus** - Metrics collection and alerting
- **Grafana** - Monitoring dashboards and visualization
- **Loki** - Log aggregation and analysis

### Infrastructure
- **Docker** - Containerization and deployment
- **Docker Compose** - Orchestration and deployment

## Development

### Local Development Setup

```bash
# Start infrastructure services
docker-compose up -d postgres nats consul minio redis

# Run auth service
cd auth-service
pip install -r requirements.txt
uvicorn app.main:app --reload --port 8090

# Run other services (in separate terminals)
cd billing-payment-service && go run cmd/main.go
cd provider-registry-service && go run cmd/main.go  
cd storage-service && go run cmd/main.go
cd scheduler-orchestrator-service && go run cmd/main.go
cd api-gateway && go run cmd/main.go

# Run frontend
cd frontend/web-app
npm install
npm run dev
```

### Service Documentation

Each service includes comprehensive documentation:

- [API Gateway](./api-gateway/README.md) - Routing, authentication, billing endpoints
- [Auth Service](./auth-service/README.md) - User management and JWT authentication  
- [Billing Payment Service](./billing-payment-service/README.md) - dGPU token payments and blockchain
- [Provider Registry Service](./provider-registry-service/README.md) - GPU provider management
- [Storage Service](./storage-service/README.md) - File storage and management
- [Scheduler Orchestrator Service](./scheduler-orchestrator-service/README.md) - Job scheduling and execution
- [Frontend Web App](./frontend/web-app/README.md) - User interface and components

## API Documentation

### Authentication Endpoints
```bash
POST /auth/login          # User login
POST /auth/register       # User registration  
GET  /auth/profile        # Get user profile
```

### Job Management Endpoints
```bash
POST /api/v1/jobs         # Submit GPU rental job
GET  /api/v1/jobs/{id}    # Get job status
DELETE /api/v1/jobs/{id}  # Cancel job
```

### Billing & Marketplace Endpoints
```bash
GET  /api/v1/billing/marketplace    # Browse available GPUs
POST /api/v1/billing/estimate       # Estimate job cost
GET  /api/v1/billing/wallet/{id}    # Get wallet balance
POST /api/v1/billing/deposit        # Deposit tokens
```

### Provider Management Endpoints
```bash
POST /api/v1/providers              # Register GPU provider
GET  /api/v1/providers              # List providers
PUT  /api/v1/providers/{id}/status  # Update provider status
```

## Deployment Options

### 🐳 Docker Compose (Recommended)

Complete platform deployment with single command:

```bash
./deploy-production.sh
```

### ☸️ Kubernetes

Deploy to Kubernetes cluster:

```bash
# Apply Kubernetes manifests
kubectl apply -f k8s/

# Check deployment status
kubectl get pods -n dante-gpu
```

### ☁️ Cloud Deployment

#### AWS ECS
```bash
# Deploy to AWS ECS
aws ecs create-cluster --cluster-name dante-gpu
# Configure task definitions and services
```

#### Google Cloud Run
```bash
# Deploy services to Cloud Run
gcloud run deploy dante-api-gateway --image gcr.io/project/api-gateway
```

#### Azure Container Instances
```bash
# Deploy to Azure
az container create --resource-group dante-gpu --file docker-compose.yml
```

## Monitoring & Operations

### Health Checks

All services provide health check endpoints:

```bash
curl http://localhost:8080/health    # API Gateway
curl http://localhost:8090/health    # Auth Service
curl http://localhost:3000/api/health # Frontend
```

### Monitoring Stack

- **Prometheus** (http://localhost:9090) - Metrics and alerts
- **Grafana** (http://localhost:3001) - Dashboards and visualization  
- **Loki** (http://localhost:3100) - Log aggregation

### Logs

View service logs:

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f api-gateway

# Live tail
docker-compose logs -f --tail=100
```

### Backup & Recovery

```bash
# Create backup
./scripts/backup.sh

# Restore from backup
./scripts/restore.sh backup-20241201.tar.gz
```

## Security

### Authentication & Authorization
- JWT token-based authentication
- Role-based access control (RBAC)
- API rate limiting and throttling
- Request validation and sanitization

### Infrastructure Security
- Container security scanning
- Network segmentation
- Secrets management
- Regular security updates

### Blockchain Security
- Multi-signature wallets for platform funds
- Transaction verification and confirmation
- Smart contract auditing
- Private key encryption

## Performance

### Scalability
- Horizontal scaling for all services
- Load balancing with health checks
- Database connection pooling
- Caching with Redis

### Optimization
- Docker multi-stage builds
- Image optimization and compression
- Database indexing and query optimization
- CDN for static assets

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
- Join our Discord community: [discord.gg/dante-gpu](https://discord.gg/dante-gpu)
- Check the documentation wiki
- Email: support@dantegpu.com

---

**🎉 Dante GPU Rental Platform - Decentralized GPU Computing for the Future**

Built with ❤️ by the Dante GPU team
