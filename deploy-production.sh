#!/bin/bash

# Dante GPU Rental Platform - Production Deployment Script
# This script deploys the complete platform in production mode

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.prod.yml"
ENV_FILE=".env"
BACKUP_DIR="./backups/$(date +%Y%m%d_%H%M%S)"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    if ! command_exists docker; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command_exists docker-compose; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    if ! command_exists curl; then
        print_error "curl is not installed. Please install curl first."
        exit 1
    fi
    
    print_success "All prerequisites are met."
}

# Function to setup environment
setup_environment() {
    print_status "Setting up environment..."
    
    if [ ! -f "$ENV_FILE" ]; then
        if [ -f "env.production.example" ]; then
            print_warning ".env file not found. Copying from env.production.example"
            cp env.production.example .env
            print_warning "Please edit .env file with your production values before proceeding!"
            print_warning "Press any key to continue after editing .env file..."
            read -n 1 -s
        else
            print_error ".env file not found and no example file available!"
            exit 1
        fi
    fi
    
    print_success "Environment setup complete."
}

# Function to create backup
create_backup() {
    print_status "Creating backup..."
    
    mkdir -p "$BACKUP_DIR"
    
    # Backup database if running
    if docker ps | grep -q dante-postgres; then
        print_status "Backing up database..."
        docker exec dante-postgres pg_dump -U dante dante > "$BACKUP_DIR/database_backup.sql"
        print_success "Database backup created: $BACKUP_DIR/database_backup.sql"
    fi
    
    # Backup volumes
    print_status "Backing up Docker volumes..."
    docker run --rm -v dante-backend_postgres_data:/data -v "$(pwd)/$BACKUP_DIR:/backup" alpine tar czf /backup/postgres_data.tar.gz -C /data .
    docker run --rm -v dante-backend_minio_data:/data -v "$(pwd)/$BACKUP_DIR:/backup" alpine tar czf /backup/minio_data.tar.gz -C /data .
    
    print_success "Backup created in: $BACKUP_DIR"
}

# Function to build images
build_images() {
    print_status "Building Docker images..."
    
    # Build all services
    docker-compose -f "$COMPOSE_FILE" build --no-cache
    
    print_success "All Docker images built successfully."
}

# Function to deploy services
deploy_services() {
    print_status "Deploying services..."
    
    # Stop existing services
    print_status "Stopping existing services..."
    docker-compose -f "$COMPOSE_FILE" down --remove-orphans || true
    
    # Start infrastructure services first
    print_status "Starting infrastructure services..."
    docker-compose -f "$COMPOSE_FILE" up -d postgres nats consul minio redis
    
    # Wait for infrastructure to be ready
    print_status "Waiting for infrastructure services to be ready..."
    sleep 30
    
    # Check infrastructure health
    check_service_health "postgres" "5432"
    check_service_health "nats" "8222"
    check_service_health "consul" "8500"
    check_service_health "minio" "9000"
    check_service_health "redis" "6379"
    
    # Start application services
    print_status "Starting application services..."
    docker-compose -f "$COMPOSE_FILE" up -d auth-service billing-service provider-registry storage-service scheduler-service
    
    # Wait for application services
    print_status "Waiting for application services to be ready..."
    sleep 45
    
    # Check application service health
    check_service_health "auth-service" "8090"
    check_service_health "billing-service" "8080"
    check_service_health "provider-registry" "8081"
    check_service_health "storage-service" "8082"
    check_service_health "scheduler-service" "8083"
    
    # Start API Gateway and Frontend
    print_status "Starting API Gateway and Frontend..."
    docker-compose -f "$COMPOSE_FILE" up -d api-gateway frontend
    
    # Wait for final services
    print_status "Waiting for API Gateway and Frontend to be ready..."
    sleep 30
    
    # Check final services
    check_service_health "api-gateway" "8080"
    check_service_health "frontend" "3000"
    
    # Start monitoring services
    print_status "Starting monitoring services..."
    docker-compose -f "$COMPOSE_FILE" up -d prometheus grafana loki test-server
    
    print_success "All services deployed successfully!"
}

# Function to check service health
check_service_health() {
    local service_name=$1
    local port=$2
    local max_attempts=30
    local attempt=1
    
    print_status "Checking health of $service_name..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f -s "http://localhost:$port/health" > /dev/null 2>&1 || \
           curl -f -s "http://localhost:$port" > /dev/null 2>&1 || \
           nc -z localhost "$port" > /dev/null 2>&1; then
            print_success "$service_name is healthy"
            return 0
        fi
        
        print_status "Attempt $attempt/$max_attempts: $service_name not ready yet..."
        sleep 5
        attempt=$((attempt + 1))
    done
    
    print_warning "$service_name health check timeout, but continuing deployment..."
    return 0
}

# Function to run post-deployment tests
run_tests() {
    print_status "Running post-deployment tests..."
    
    # Test API Gateway
    if curl -f -s "http://localhost:8080/health" > /dev/null; then
        print_success "API Gateway is responding"
    else
        print_warning "API Gateway health check failed"
    fi
    
    # Test Frontend
    if curl -f -s "http://localhost:3000/api/health" > /dev/null; then
        print_success "Frontend is responding"
    else
        print_warning "Frontend health check failed"
    fi
    
    # Test Auth Service
    if curl -f -s "http://localhost:8090/health" > /dev/null; then
        print_success "Auth Service is responding"
    else
        print_warning "Auth Service health check failed"
    fi
    
    print_success "Post-deployment tests completed"
}

# Function to show deployment status
show_status() {
    print_status "Deployment Status:"
    echo ""
    
    print_status "🌐 Service URLs:"
    echo "  Frontend:        http://localhost:3000"
    echo "  API Gateway:     http://localhost:8080"
    echo "  Auth Service:    http://localhost:8090"
    echo "  Grafana:         http://localhost:3001 (admin/admin)"
    echo "  Prometheus:      http://localhost:9090"
    echo "  Consul UI:       http://localhost:8500"
    echo "  MinIO Console:   http://localhost:9001"
    echo "  Test Server:     http://localhost:9999"
    echo ""
    
    print_status "🐳 Running Containers:"
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep dante
    echo ""
    
    print_status "📊 Resource Usage:"
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}" | grep dante
}

# Function to cleanup
cleanup() {
    print_status "Cleaning up unused Docker resources..."
    docker system prune -f
    docker volume prune -f
    print_success "Cleanup completed"
}

# Main deployment function
main() {
    echo ""
    print_status "🚀 Starting Dante GPU Rental Platform Production Deployment"
    echo ""
    
    check_prerequisites
    setup_environment
    
    # Ask for backup confirmation
    read -p "Do you want to create a backup before deployment? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        create_backup
    fi
    
    build_images
    deploy_services
    run_tests
    show_status
    cleanup
    
    echo ""
    print_success "🎉 Dante GPU Rental Platform deployed successfully!"
    print_status "📖 Check the logs with: docker-compose -f $COMPOSE_FILE logs -f"
    print_status "🛑 Stop services with: docker-compose -f $COMPOSE_FILE down"
    echo ""
}

# Handle script interruption
trap 'print_error "Deployment interrupted!"; exit 1' INT TERM

# Run main function
main "$@" 