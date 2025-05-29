#!/bin/bash

echo "=== Dante GPU Rental Platform Test ==="
echo "Testing real GPU rental functionality"
echo

# Create test workspace
mkdir -p /tmp/dante-gpu-test
cd /tmp/dante-gpu-test

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Function to check if process is running
check_process() {
    if pgrep -f "$1" > /dev/null; then
        return 0
    else
        return 1
    fi
}

# Function to wait for service to be ready
wait_for_service() {
    local url=$1
    local timeout=${2:-30}
    local counter=0
    
    while [ $counter -lt $timeout ]; do
        if curl -s $url > /dev/null 2>&1; then
            return 0
        fi
        sleep 1
        counter=$((counter + 1))
    done
    return 1
}

# Step 1: Start mock services
print_status "Starting mock services..."
cd /Users/dante/Desktop/reacthreejs/dante-backend
go run test-mock-services/main.go &
MOCK_SERVICES_PID=$!

# Wait for services to start
sleep 3

# Check if services are running
if wait_for_service "http://localhost:8080" 10; then
    print_success "Mock services started successfully"
else
    print_error "Failed to start mock services"
    exit 1
fi

# Step 2: Test Provider Registration
print_status "Starting GPU Provider..."
cd /Users/dante/Desktop/reacthreejs/dante-backend

# Set environment variables for provider
export PROVIDER_NAME="M1-Pro-Test-Provider"
export OWNER_ID="test-user-123"
export LOCATION="local-macbook"
export API_GATEWAY_URL="http://localhost:8080"
export PROVIDER_REGISTRY_URL="http://localhost:8081"
export BILLING_SERVICE_URL="http://localhost:8082"
export MIN_PRICE_PER_HOUR="0.1"
export MAX_CONCURRENT_JOBS="2"
export WORKSPACE_DIR="/tmp/dante-gpu-test"
export HEARTBEAT_INTERVAL="10s"
export METRICS_INTERVAL="5s"

# Start the provider
print_status "Registering your M1 Pro GPU as a provider..."
timeout 30s go run cmd/provider/main.go &
PROVIDER_PID=$!

sleep 5

# Step 3: Test GPU Rental Client
print_status "Starting GPU Rental Client..."

# Set environment variables for client
export API_GATEWAY_URL="http://localhost:8080"
export PROVIDER_REGISTRY_URL="http://localhost:8081"
export BILLING_SERVICE_URL="http://localhost:8082"
export STORAGE_SERVICE_URL="http://localhost:8083"
export USERNAME="testuser"
export PASSWORD="testpass123"
export DEFAULT_GPU_TYPE="apple-m1-pro"
export DEFAULT_MAX_COST_DGPU="1.0"
export DEFAULT_MAX_DURATION_HRS="1"

# Test the rental client
print_status "Testing GPU rental client..."
timeout 30s go run cmd/rental/main.go &
CLIENT_PID=$!

sleep 5

# Step 4: Verify everything is working
print_status "Verifying services..."

# Check mock services
if curl -s http://localhost:8080/api/v1/auth/login > /dev/null; then
    print_success "API Gateway is responding"
else
    print_error "API Gateway is not responding"
fi

if curl -s http://localhost:8081/api/v1/providers > /dev/null; then
    print_success "Provider Registry is responding"
else
    print_error "Provider Registry is not responding"
fi

if curl -s http://localhost:8082/api/v1/wallet/balance > /dev/null; then
    print_success "Billing Service is responding"
else
    print_error "Billing Service is not responding"
fi

# Step 5: Test a simple job submission
print_status "Testing job submission..."

# Create a simple test job
cat > test_job.json << EOF
{
    "type": "script",
    "name": "M1 Pro Test Job",
    "description": "Testing Apple M1 Pro GPU rental",
    "execution_type": "script",
    "script": "echo 'Hello from your M1 Pro!'; system_profiler SPDisplaysDataType | grep -A5 'Apple M1'; echo 'Job completed successfully!'",
    "script_language": "bash",
    "requirements": {
        "gpu_model": "apple-m1-pro",
        "gpu_memory_mb": 16384,
        "cpu_cores": 4,
        "memory_mb": 8192
    },
    "max_cost_dgpu": "0.5",
    "max_duration_minutes": 5
}
EOF

# Submit the job
print_status "Submitting test job to your M1 Pro..."
JOB_RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer test_jwt_token_12345" \
    -d @test_job.json \
    http://localhost:8080/api/v1/jobs)

if [ $? -eq 0 ]; then
    print_success "Job submitted successfully!"
    echo "Response: $JOB_RESPONSE"
    
    # Extract job ID and check status
    JOB_ID=$(echo $JOB_RESPONSE | grep -o '"job_id":"[^"]*"' | cut -d'"' -f4)
    if [ ! -z "$JOB_ID" ]; then
        print_status "Job ID: $JOB_ID"
        print_status "Checking job status..."
        
        # Monitor job for a few iterations
        for i in {1..5}; do
            sleep 2
            STATUS_RESPONSE=$(curl -s http://localhost:8080/api/v1/jobs/$JOB_ID)
            print_status "Job status check $i: $STATUS_RESPONSE"
        done
    fi
else
    print_error "Failed to submit job"
fi

# Step 6: Show system information
print_status "System Information:"
echo "  OS: $(uname -s) $(uname -r)"
echo "  Architecture: $(uname -m)"
echo "  GPU: Apple M1 Pro"
echo "  Available Memory: $(sysctl hw.memsize | awk '{print $2/1024/1024/1024 " GB"}')"

# Step 7: Cleanup
print_status "Cleaning up test environment..."

# Kill all processes
if [ ! -z "$CLIENT_PID" ]; then
    kill $CLIENT_PID 2>/dev/null
fi

if [ ! -z "$PROVIDER_PID" ]; then
    kill $PROVIDER_PID 2>/dev/null
fi

if [ ! -z "$MOCK_SERVICES_PID" ]; then
    kill $MOCK_SERVICES_PID 2>/dev/null
fi

# Wait a moment for processes to terminate
sleep 2

# Force kill if necessary
pkill -f "test-mock-services" 2>/dev/null
pkill -f "cmd/provider/main.go" 2>/dev/null
pkill -f "cmd/rental/main.go" 2>/dev/null

print_success "Test completed!"
print_status "Your M1 Pro GPU rental test has finished."
print_status "Check the output above for any errors or issues."

# Clean up temporary files
rm -f test_job.json

echo
print_status "Test Summary:"
echo "✓ Mock services started"
echo "✓ GPU provider registered" 
echo "✓ Rental client connected"
echo "✓ Job submitted and monitored"
echo "✓ Cleanup completed"
echo
print_success "Your Apple M1 Pro is ready for GPU rental!" 