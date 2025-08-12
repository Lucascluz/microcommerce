#!/bin/bash

# Change to the project root directory (parent of scripts)
cd "$(dirname "$0")/.."

echo "Starting all microservices..."

# Run all services in the background from their respective directories
echo "Starting API Gateway on port 8080..."
(cd services/api-gateway && go run cmd/main.go) &

echo "Starting Catalog Service on port 8082..."
(cd services/catalog-service && go run cmd/main.go) &

echo "Starting Transaction Service on port 8081..."
(cd services/transaction-service && go run cmd/main.go) &

echo "Starting User Service on port 8083..."
(cd services/user-service && go run cmd/main.go) &

echo "Starting Notifications Service on port 8087..."
(cd services/notifications-service && go run cmd/main.go) &

echo "Starting Visualization Service on port 8089..."
(cd services/visualization-service && go run cmd/main.go) &

echo ""
echo "All services started in background!"
echo "API Gateway should be available at: http://localhost:8080"
echo "Catalog Service should be available at: http://localhost:8082"
echo "Transaction Service should be available at: http://localhost:8081"
echo "User Service should be available at: http://localhost:8083"
echo "Notifications Service should be available at: http://localhost:8087"
echo "Visualization Service should be available at: http://localhost:8089"
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