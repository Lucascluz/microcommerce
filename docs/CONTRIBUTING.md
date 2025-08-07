# Contributing to MicroCommerce

Thank you for considering contributing to MicroCommerce! This document provides guidelines and information for contributors.

## 📋 Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Getting Started](#getting-started)
3. [Development Process](#development-process)
4. [Coding Standards](#coding-standards)
5. [Testing Guidelines](#testing-guidelines)
6. [Submitting Changes](#submitting-changes)
7. [Issue Reporting](#issue-reporting)
8. [Feature Requests](#feature-requests)
9. [Documentation](#documentation)
10. [Community](#community)

## 🤝 Code of Conduct

### Our Pledge

We are committed to making participation in this project a harassment-free experience for everyone, regardless of age, body size, disability, ethnicity, gender identity and expression, level of experience, nationality, personal appearance, race, religion, or sexual identity and orientation.

### Our Standards

**Positive behavior includes:**
- Using welcoming and inclusive language
- Being respectful of differing viewpoints and experiences
- Gracefully accepting constructive criticism
- Focusing on what is best for the community
- Showing empathy towards other community members

**Unacceptable behavior includes:**
- The use of sexualized language or imagery
- Trolling, insulting/derogatory comments, and personal or political attacks
- Public or private harassment
- Publishing others' private information without explicit permission
- Other conduct which could reasonably be considered inappropriate in a professional setting

## 🚀 Getting Started

### Prerequisites

Before contributing, ensure you have:

- Go 1.22 or higher installed
- Docker and Docker Compose
- Kubernetes cluster (minikube for local development)
- Git configured with your GitHub account
- Basic understanding of microservices architecture

### Setting Up Development Environment

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/your-username/microcommerce.git
   cd microcommerce
   ```

3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/original-owner/microcommerce.git
   ```

4. **Set up the development environment**:
   ```bash
   # Start local development
   ./scripts/run-all.sh
   
   # Or use Kubernetes with Tilt
   minikube start
   tilt up
   ```

5. **Verify everything works**:
   ```bash
   curl http://localhost:8080/api/v1/services/health
   ```

### Project Structure Understanding

Familiarize yourself with the project structure:

```
microcommerce/
├── docs/                    # Documentation
├── services/               # Microservices
│   ├── api-gateway/       # API Gateway service
│   ├── payment-service/   # Payment processing
│   ├── product-service/   # Product management
│   └── user-service/      # User management
├── shared/                # Shared utilities
├── k8s/                   # Kubernetes manifests
├── scripts/               # Utility scripts
└── Tiltfile              # Tilt configuration
```

## 🔄 Development Process

### Branch Strategy

We use **GitFlow** branching model:

- `main`: Production-ready code
- `develop`: Integration branch for features
- `feature/*`: Feature development branches
- `hotfix/*`: Critical bug fixes
- `release/*`: Release preparation branches

### Workflow

1. **Create a feature branch** from `develop`:
   ```bash
   git checkout develop
   git pull upstream develop
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following our coding standards

3. **Test your changes** thoroughly:
   ```bash
   # Run tests
   go test ./...
   
   # Test integration
   curl http://localhost:8080/api/v1/services/health
   ```

4. **Commit your changes** with descriptive messages:
   ```bash
   git add .
   git commit -m "feat: add user authentication endpoint"
   ```

5. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Create a Pull Request** to the `develop` branch

### Commit Message Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(payment): add credit card validation
fix(api-gateway): resolve service discovery issue
docs: update API documentation for user endpoints
test(product): add unit tests for product service
```

## 📝 Coding Standards

### Go Code Style

Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and these additional guidelines:

#### File Organization

```go
// Package declaration
package main

// Imports (standard library first, then third-party, then local)
import (
    "context"
    "fmt"
    "log"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/segmentio/kafka-go"

    "github.com/lucas/shared/utils"
)

// Constants
const (
    DefaultPort = "8080"
    MaxRetries  = 3
)

// Variables
var (
    logger *log.Logger
    config *Config
)

// Types
type Service struct {
    // fields
}

// Functions
func main() {
    // implementation
}
```

#### Naming Conventions

- **Packages**: lowercase, single word when possible
- **Files**: lowercase with underscores (`user_service.go`)
- **Functions/Methods**: CamelCase for exported, camelCase for private
- **Variables**: camelCase
- **Constants**: CamelCase or ALL_CAPS for package-level

#### Error Handling

```go
// Good: Wrap errors with context
func processPayment(id string) error {
    payment, err := repo.GetPayment(id)
    if err != nil {
        return fmt.Errorf("failed to get payment %s: %w", id, err)
    }
    
    if err := validatePayment(payment); err != nil {
        return fmt.Errorf("payment validation failed: %w", err)
    }
    
    return nil
}

// Good: Handle errors at the appropriate level
func handlePayment(c *gin.Context) {
    if err := processPayment(c.Param("id")); err != nil {
        log.Printf("Payment processing failed: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Payment processing failed",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"status": "success"})
}
```

#### Documentation

Document all exported functions, types, and packages:

```go
// PaymentService handles payment processing operations.
// It provides methods for creating, validating, and processing payments.
type PaymentService struct {
    repo   PaymentRepository
    logger *log.Logger
}

// ProcessPayment processes a payment transaction.
// It validates the payment data, charges the payment method,
// and returns the transaction result.
//
// Parameters:
//   - ctx: Context for the operation
//   - payment: Payment data to process
//
// Returns:
//   - *PaymentResult: The result of the payment processing
//   - error: Any error that occurred during processing
func (s *PaymentService) ProcessPayment(ctx context.Context, payment *Payment) (*PaymentResult, error) {
    // implementation
}
```

### Code Quality Tools

Use these tools to maintain code quality:

```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Vet for issues
go vet ./...

# Security check
gosec ./...

# Check for ineffective assignments
ineffassign ./...

# Find unused code
deadcode ./...
```

## 🧪 Testing Guidelines

### Test Structure

Organize tests alongside your code:

```
services/payment-service/
├── cmd/
├── handlers/
│   ├── payment.go
│   └── payment_test.go
├── models/
│   ├── payment.go
│   └── payment_test.go
└── services/
    ├── payment.go
    └── payment_test.go
```

### Test Types

#### Unit Tests

Test individual functions and methods:

```go
func TestValidatePayment(t *testing.T) {
    tests := []struct {
        name    string
        payment *Payment
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid payment",
            payment: &Payment{
                Amount:   100.0,
                Currency: "USD",
                Method:   "card",
            },
            wantErr: false,
        },
        {
            name: "invalid amount",
            payment: &Payment{
                Amount:   -100.0,
                Currency: "USD",
                Method:   "card",
            },
            wantErr: true,
            errMsg:  "amount must be positive",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidatePayment(tt.payment)
            
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

#### Integration Tests

Test service interactions:

```go
func TestPaymentServiceIntegration(t *testing.T) {
    // Setup test server
    gin.SetMode(gin.TestMode)
    router := setupRouter()
    
    // Test data
    payment := map[string]interface{}{
        "amount":   100.0,
        "currency": "USD",
        "method":   "card",
    }
    
    paymentJSON, _ := json.Marshal(payment)
    
    // Make request
    req, _ := http.NewRequest("POST", "/payments", bytes.NewBuffer(paymentJSON))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // Assertions
    assert.Equal(t, http.StatusOK, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, "success", response["status"])
}
```

#### End-to-End Tests

Test complete user workflows:

```go
func TestCompletePaymentFlow(t *testing.T) {
    // This would test the entire flow from API Gateway to Payment Service
    // Including Kafka messaging
}
```

### Test Coverage

Maintain high test coverage:

```bash
# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Aim for at least 80% coverage
```

### Mocking

Use interfaces for testing:

```go
// PaymentRepository interface for mocking
type PaymentRepository interface {
    Create(ctx context.Context, payment *Payment) error
    GetByID(ctx context.Context, id string) (*Payment, error)
}

// Mock implementation
type MockPaymentRepository struct {
    payments map[string]*Payment
}

func (m *MockPaymentRepository) Create(ctx context.Context, payment *Payment) error {
    m.payments[payment.ID] = payment
    return nil
}

func (m *MockPaymentRepository) GetByID(ctx context.Context, id string) (*Payment, error) {
    payment, exists := m.payments[id]
    if !exists {
        return nil, errors.New("payment not found")
    }
    return payment, nil
}
```

## 📤 Submitting Changes

### Pull Request Process

1. **Ensure your code follows** our coding standards
2. **Add tests** for new functionality
3. **Update documentation** if needed
4. **Ensure all tests pass**:
   ```bash
   go test ./...
   ```

5. **Update the CHANGELOG** if your changes are user-facing
6. **Create a Pull Request** with:
   - Clear title and description
   - Reference to related issues
   - Screenshots/demos if applicable

### Pull Request Template

```markdown
## Description
Brief description of the changes.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Related Issues
Fixes #(issue number)

## How Has This Been Tested?
- [ ] Unit tests
- [ ] Integration tests
- [ ] Manual testing

## Screenshots (if applicable)

## Checklist
- [ ] My code follows the style guidelines of this project
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
```

### Review Process

1. **Automated checks** must pass (CI/CD pipeline)
2. **Code review** by at least one maintainer
3. **Testing** in staging environment (for significant changes)
4. **Approval** from maintainer
5. **Merge** to develop branch

## 🐛 Issue Reporting

### Before Reporting

1. **Search existing issues** to avoid duplicates
2. **Try the latest version** to see if it's already fixed
3. **Check documentation** for possible solutions

### Bug Report Template

```markdown
## Bug Description
A clear and concise description of what the bug is.

## Steps to Reproduce
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

## Expected Behavior
A clear and concise description of what you expected to happen.

## Actual Behavior
A clear and concise description of what actually happened.

## Environment
- OS: [e.g. Ubuntu 20.04]
- Go version: [e.g. 1.22.0]
- Kubernetes version: [e.g. 1.25.0]
- Docker version: [e.g. 20.10.0]

## Additional Context
Add any other context about the problem here.

## Logs
```
Paste relevant logs here
```
```

## 💡 Feature Requests

### Feature Request Template

```markdown
## Feature Summary
A clear and concise description of the feature you'd like to see.

## Problem Statement
What problem does this feature solve?

## Proposed Solution
Describe the solution you'd like to see implemented.

## Alternatives Considered
Describe any alternative solutions or features you've considered.

## Use Cases
Describe specific use cases where this feature would be beneficial.

## Implementation Details
If you have ideas about how this could be implemented, please share them.

## Additional Context
Add any other context or screenshots about the feature request here.
```

### Feature Development Process

1. **Discussion** in GitHub issues
2. **Design document** for significant features
3. **Implementation plan** agreement
4. **Development** in feature branch
5. **Testing** and review
6. **Documentation** updates
7. **Release** planning

## 📚 Documentation

### Documentation Types

- **API Documentation**: OpenAPI/Swagger specs
- **Architecture Documentation**: System design and patterns
- **User Documentation**: Installation and usage guides
- **Developer Documentation**: Contributing and development guides

### Documentation Guidelines

- Use clear, concise language
- Include code examples
- Keep documentation up-to-date with code changes
- Use proper Markdown formatting
- Include diagrams where helpful

### Documentation Structure

```markdown
# Title

## Overview
Brief description of what this document covers.

## Table of Contents
1. [Section 1](#section-1)
2. [Section 2](#section-2)

## Section 1
Content with code examples:

```go
// Code example
func example() {
    // Implementation
}
```

## Examples
Practical examples and use cases.

## References
Links to related documentation.
```

## 🌟 Community

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and community discussions
- **Pull Requests**: Code reviews and discussions
- **Wiki**: Additional documentation and guides

### Community Guidelines

- Be respectful and inclusive
- Help newcomers get started
- Share knowledge and best practices
- Provide constructive feedback
- Celebrate contributions from all community members

### Recognition

We recognize contributors in several ways:

- **Contributors file**: All contributors are listed
- **Release notes**: Significant contributions are highlighted
- **Community showcases**: Featuring interesting use cases and implementations

## 📄 License

By contributing to MicroCommerce, you agree that your contributions will be licensed under the same license as the project (MIT License).

## ❓ Questions?

If you have questions about contributing:

1. Check the [FAQ](docs/FAQ.md)
2. Search [existing issues](https://github.com/owner/microcommerce/issues)
3. Create a new [discussion](https://github.com/owner/microcommerce/discussions)
4. Reach out to maintainers

Thank you for contributing to MicroCommerce! 🎉
