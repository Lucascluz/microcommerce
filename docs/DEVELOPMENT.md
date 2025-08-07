# Development Guide

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Environment Setup](#environment-setup)
3. [Local Development](#local-development)
4. [Development Workflow](#development-workflow)
5. [Code Standards](#code-standards)
6. [Testing Guidelines](#testing-guidelines)
7. [Debugging](#debugging)
8. [Performance Optimization](#performance-optimization)
9. [Common Issues](#common-issues)

## Prerequisites

### System Requirements

- **Go**: Version 1.22 or higher
- **Docker**: Version 20.0 or higher
- **Docker Compose**: Version 2.0 or higher
- **Kubernetes**: minikube, kind, or Docker Desktop
- **Git**: Version 2.0 or higher

### Optional Tools

- **Tilt**: For enhanced Kubernetes development experience
- **kubectl**: Kubernetes CLI tool
- **Helm**: Kubernetes package manager
- **Postman**: API testing tool
- **Visual Studio Code**: Recommended IDE with Go extension

### Verify Installation

```bash
# Check Go version
go version

# Check Docker version  
docker --version

# Check Kubernetes
kubectl version --client

# Check Tilt (if installed)
tilt version
```

## Environment Setup

### 1. Clone Repository

```bash
git clone <repository-url>
cd microcommerce
```

### 2. Go Module Setup

The project uses Go modules with local replace directives. No additional setup required for dependencies.

```bash
# Verify module structure
cd services/api-gateway && go mod tidy
cd ../payment-service && go mod tidy
cd ../product-service && go mod tidy
cd ../user-service && go mod tidy
cd ../../shared && go mod tidy
```

### 3. Environment Variables

Create a `.env` file in the project root (optional):

```bash
# .env
KAFKA_BROKER=localhost:9092
API_GATEWAY_PORT=8080
PAYMENT_SERVICE_PORT=8081
PRODUCT_SERVICE_PORT=8082
USER_SERVICE_PORT=8083
```

## Local Development

### Option 1: Native Go Development

Start services directly with Go:

```bash
# Terminal 1: Start API Gateway
cd services/api-gateway
go run cmd/main.go

# Terminal 2: Start Payment Service
cd services/payment-service
go run cmd/main.go

# Terminal 3: Start Product Service  
cd services/product-service
go run cmd/main.go

# Terminal 4: Start User Service
cd services/user-service
go run cmd/main.go
```

Or use the convenience script:

```bash
# Start all services
chmod +x scripts/run-all.sh
./scripts/run-all.sh

# Stop all services
./scripts/stop-all.sh
```

### Option 2: Kubernetes with Tilt (Recommended)

```bash
# Start minikube
minikube start

# Start development environment
tilt up

# View Tilt dashboard
open http://localhost:10350
```

### Option 3: Docker Compose (Planned)

```bash
# Start with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

## Development Workflow

### 1. Feature Development

```bash
# Create feature branch
git checkout -b feature/user-authentication

# Make changes
# ... code changes ...

# Test locally
go test ./...

# Commit changes
git add .
git commit -m "feat: implement user authentication"

# Push branch
git push origin feature/user-authentication
```

### 2. Adding New Endpoints

1. **Define the endpoint** in the appropriate service
2. **Update API documentation** in `docs/API.md`
3. **Add tests** for the new functionality
4. **Update health checks** if needed

Example: Adding a new endpoint to Payment Service

```go
// services/payment-service/cmd/main.go
func main() {
    // ... existing code ...
    
    // Add new endpoint
    router.POST("/payments", handlePayment)
    
    // ... rest of code ...
}

func handlePayment(c *gin.Context) {
    // Implementation
    c.JSON(http.StatusOK, gin.H{"status": "payment processed"})
}
```

### 3. Adding New Services

1. **Create service directory** under `services/`
2. **Copy structure** from existing service
3. **Update go.mod** with appropriate module name
4. **Add Dockerfile** for containerization
5. **Create Kubernetes manifests** under `k8s/`
6. **Update Tiltfile** to include new service
7. **Update scripts** to include new service

## Code Standards

### Go Code Style

Follow the official Go style guide and use these tools:

```bash
# Format code
go fmt ./...

# Lint code (install golangci-lint first)
golangci-lint run

# Vet code for issues
go vet ./...
```

### File Structure

```
services/[service-name]/
├── cmd/
│   └── main.go          # Entry point
├── handlers/            # HTTP handlers
│   └── handlers.go
├── models/              # Data models
│   └── models.go
├── services/            # Business logic
│   └── service.go
├── repository/          # Data access layer
│   └── repository.go
├── middleware/          # HTTP middleware
│   └── middleware.go
├── config/              # Configuration
│   └── config.go
├── Dockerfile
├── go.mod
└── go.sum
```

### Naming Conventions

- **Packages**: lowercase, single word when possible
- **Files**: lowercase with underscores
- **Functions**: CamelCase (exported) or camelCase (private)
- **Variables**: camelCase
- **Constants**: ALL_CAPS or CamelCase for exported

### Error Handling

```go
// Good: Wrap errors with context
func processPayment(id string) error {
    payment, err := repo.GetPayment(id)
    if err != nil {
        return fmt.Errorf("failed to get payment %s: %w", id, err)
    }
    
    // Process payment...
    return nil
}

// Good: Handle errors at appropriate level
func handler(c *gin.Context) {
    err := processPayment(c.Param("id"))
    if err != nil {
        log.Printf("Payment processing failed: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Payment processing failed",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"status": "success"})
}
```

### Logging

Use structured logging:

```go
import "log"

// Good: Structured logging with context
log.Printf("Payment processed: id=%s, amount=%.2f, status=%s", 
    payment.ID, payment.Amount, payment.Status)

// Better: Use structured logging library (planned)
logger.Info("Payment processed",
    "payment_id", payment.ID,
    "amount", payment.Amount,
    "status", payment.Status,
)
```

## Testing Guidelines

### Unit Tests

Create tests alongside your code:

```go
// services/payment-service/handlers/handlers_test.go
package handlers

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func TestPaymentHandler(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    router := gin.New()
    router.POST("/payments", handlePayment)
    
    req, _ := http.NewRequest("POST", "/payments", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
}
```

### Integration Tests

Test service interactions:

```go
func TestServiceHealth(t *testing.T) {
    // Start test server
    // Make HTTP request
    // Verify response
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v ./services/payment-service/handlers -run TestPaymentHandler
```

### Test Data

Use table-driven tests:

```go
func TestValidatePayment(t *testing.T) {
    tests := []struct {
        name    string
        payment Payment
        wantErr bool
    }{
        {
            name: "valid payment",
            payment: Payment{Amount: 100.0, Currency: "USD"},
            wantErr: false,
        },
        {
            name: "invalid amount",
            payment: Payment{Amount: -100.0, Currency: "USD"},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidatePayment(tt.payment)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidatePayment() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Debugging

### Local Debugging

#### Using VS Code

1. Install Go extension
2. Set breakpoints in code
3. Run with debugger (F5)

Debug configuration (`.vscode/launch.json`):

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug API Gateway",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/services/api-gateway/cmd/main.go",
            "cwd": "${workspaceFolder}/services/api-gateway"
        }
    ]
}
```

#### Using Delve

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug service
cd services/api-gateway
dlv debug cmd/main.go

# Set breakpoint and continue
(dlv) break main.main
(dlv) continue
```

### Kubernetes Debugging

#### View Logs

```bash
# View pod logs
kubectl logs -f deployment/api-gateway

# View all logs for a service
kubectl logs -f -l app=api-gateway

# View logs with Tilt
tilt logs api-gateway
```

#### Debug Pod Issues

```bash
# Describe pod for events
kubectl describe pod <pod-name>

# Execute into pod
kubectl exec -it <pod-name> -- /bin/sh

# Port forward for debugging
kubectl port-forward pod/<pod-name> 8080:8080
```

### Kafka Debugging

#### Check Kafka Connection

```bash
# Test Kafka connection
kafkacat -b localhost:9092 -L

# Consume messages from topic
kafkacat -b localhost:9092 -t service-ping -C

# Produce test message
echo "test message" | kafkacat -b localhost:9092 -t service-ping -P
```

## Performance Optimization

### Profiling

Enable pprof for performance profiling:

```go
import _ "net/http/pprof"

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // ... rest of application
}
```

Access profiles at:
- `http://localhost:6060/debug/pprof/`
- `http://localhost:6060/debug/pprof/heap`
- `http://localhost:6060/debug/pprof/profile`

### Memory Optimization

```bash
# Check memory usage
go tool pprof http://localhost:6060/debug/pprof/heap

# CPU profiling
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

### Performance Testing

Use `hey` for load testing:

```bash
# Install hey
go install github.com/rakyll/hey@latest

# Load test API
hey -n 1000 -c 10 http://localhost:8080/api/v1/services/health
```

## Common Issues

### Port Already in Use

```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>
```

### Go Module Issues

```bash
# Clean module cache
go clean -modcache

# Refresh dependencies
go mod tidy
go mod download
```

### Kafka Connection Issues

```bash
# Check if Kafka is running
docker ps | grep kafka

# Restart Kafka in Kubernetes
kubectl delete pod -l app=kafka
```

### Service Discovery Issues

```bash
# Check service endpoints
kubectl get endpoints

# Verify service DNS
kubectl exec -it <pod-name> -- nslookup kafka
```

## IDE Configuration

### VS Code Extensions

Recommended extensions:
- Go (golang.go)
- Kubernetes (ms-kubernetes-tools.vscode-kubernetes-tools)
- Docker (ms-azuretools.vscode-docker)
- YAML (redhat.vscode-yaml)
- GitLens (eamodio.gitlens)

### VS Code Settings

```json
{
    "go.useLanguageServer": true,
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.testFlags": ["-v"],
    "go.buildTags": "integration",
    "editor.formatOnSave": true
}
```

## Git Hooks

Set up pre-commit hooks:

```bash
# .git/hooks/pre-commit
#!/bin/sh
go fmt ./...
go vet ./...
go test ./...
```

```bash
# Make executable
chmod +x .git/hooks/pre-commit
```

## Documentation Updates

When making changes:

1. **Update API docs** if adding/changing endpoints
2. **Update architecture docs** if changing system design
3. **Update README** if changing setup/deployment
4. **Add inline comments** for complex business logic

## Next Steps

After setting up development environment:

1. **Implement business logic** for each service
2. **Add database integration** (PostgreSQL/MongoDB)
3. **Implement authentication** system
4. **Add comprehensive testing**
5. **Set up CI/CD pipeline**
