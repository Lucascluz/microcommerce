#!/bin/bash

# Change to the project root directory (parent of scripts)
cd "$(dirname "$0")/.."

echo "Starting all microservices..."

# Run all services in the background from their respective directories
echo "Starting API Gateway on port 8080..."
(cd services/api-gateway && go run cmd/main.go) &

echo "Starting Payment Service..."
(cd services/payment-service && go run cmd/main.go) &

echo "Starting Product Service..."
(cd services/product-service && go run cmd/main.go) &

echo "Starting User Service..."
(cd services/user-service && go run cmd/main.go) &

echo ""
echo "All services started in background!"
echo "API Gateway should be available at: http://localhost:8080"
echo "Press Ctrl+C to stop all services"
echo ""

# Function to cleanup background processes on exit
cleanup() {
    echo "Stopping all services..."
    pkill -f "go run services/"
    exit 0
}

# Set trap to cleanup on Ctrl+C
trap cleanup INT

# Wait for all background processes to finish
wait