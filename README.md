# MicroCommerce - Restructured Architecture

A distributed e-commerce microservices platform built with Go, designed for scalability and maintainability.

## 🏗️ Architecture Overview

The system has been restructured into **5 core services** for better organization and reduced complexity:

### Core Services

| Service | Port | Responsibility | Consolidated From |
|---------|------|---------------|-------------------|
| **api-gateway** | 8080 | Request routing, service discovery | - |
| **user-service** | 8083 | Authentication, profiles, account management | *(unchanged)* |
| **catalog-service** | 8082 | Product listings, reviews, inventory | product-service + review-service |
| **transaction-service** | 8081 | Orders, payments, sales, shipping | payment-service + order-service + sales-service + shipping-service |
| **notification-service** | 8087 | Email, push notification | *(unchanged)* |
| **visualization-service** | 8089 | Analytics, observability, reports | *(unchanged)* |

### Shared Infrastructure

| Component | Port | Purpose |
|-----------|------|---------|
| **PostgreSQL** | 5432 | Primary database for all services |
| **Redis** | 6379 | Caching and session management |
| **Kafka** | 9092 | Inter-service messaging |

### Architecture Diagram

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Client    │    │  Load Bal.  │    │   API GW    │
│ (Frontend)  │◄──►│  (Optional) │◄──►│  (Port 8080)│
└─────────────┘    └─────────────┘    └─────────────┘
                                             │
                   ┌─────────────────────────┼─────────────────────────┐
                   │                         │                         │
                   ▼                         ▼                         ▼
            ┌─────────────┐          ┌─────────────┐          ┌─────────────┐
            │   Payment   │          │   Product   │          │    User     │
            │   Service   │          │   Service   │          │   Service   │
            │ (Port 8081) │          │ (Port 8082) │          │ (Port 8083) │
            └─────────────┘          └─────────────┘          └─────────────┘
                   │                         │                         │
                   └─────────────────────────┼─────────────────────────┘
                                             │
                                             ▼
                                    ┌─────────────┐
                                    │    Kafka    │
                                    │ (Port 9092) │
                                    └─────────────┘
```

## 🛠️ Technology Stack

- **Backend**: Go 1.22+ with Gin web framework
- **Message Broker**: Apache Kafka 3.7.0
- **Container Runtime**: Docker
- **Orchestration**: Kubernetes
- **Development Tool**: Tilt for local development
- **Service Mesh**: Ready for Istio integration

## 📋 Prerequisites

- Go 1.22 or higher
- Docker and Docker Compose
- Kubernetes cluster (minikube for local development)
- Tilt (optional, for enhanced development experience)

## 🚀 Quick Start

### Option 1: Local Development (Native)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd microcommerce
   ```

2. **Start all services**
   ```bash
   chmod +x scripts/run-all.sh
   ./scripts/run-all.sh
   ```

3. **Verify services are running**
   ```bash
   curl http://localhost:8080/api/v1/services/health
   ```

4. **Stop all services**
   ```bash
   ./scripts/stop-all.sh
   ```

### Option 2: Kubernetes with Tilt (Recommended)

1. **Start minikube**
   ```bash
   minikube start
   ```

2. **Run with Tilt**
   ```bash
   tilt up
   ```

3. **Access the Tilt dashboard**
   Open http://localhost:10350 in your browser

## 📚 Documentation

- [**Architecture Guide**](docs/ARCHITECTURE.md) - Detailed system design and patterns
- [**API Documentation**](docs/API.md) - Complete API reference
- [**Deployment Guide**](docs/DEPLOYMENT.md) - Production deployment instructions
- [**Development Guide**](docs/DEVELOPMENT.md) - Local development setup and guidelines
- [**Contributing**](docs/CONTRIBUTING.md) - How to contribute to the project
- [**Troubleshooting**](docs/TROUBLESHOOTING.md) - Common issues and solutions

## 🔌 Service Endpoints

| Service | Port | Health Check | Description |
|---------|------|--------------|-------------|
| API Gateway | 8080 | `GET /` | Main entry point and service orchestration |
| Payment Service | 8081 | `GET /` | Payment processing and transaction management |
| Product Service | 8082 | `GET /` | Product catalog and inventory management |
| User Service | 8083 | `GET /` | User authentication and profile management |

## 🔄 Event-Driven Communication

Services communicate through Kafka topics:

- **service-ping**: API Gateway sends health checks
- **service-pong**: Services respond with their status
- Additional topics for business events (orders, payments, inventory updates)

## 🏗️ Project Structure

```
microcommerce/
├── docs/                      # Documentation
├── k8s/                       # Kubernetes manifests
│   ├── api-gateway/
│   ├── payment-service/
│   ├── product-service/
│   ├── user-service/
│   └── kafka/
├── scripts/                   # Utility scripts
├── services/                  # Microservices
│   ├── api-gateway/
│   ├── payment-service/
│   ├── product-service/
│   └── user-service/
├── shared/                    # Shared utilities
├── Tiltfile                   # Tilt configuration
└── README.md
```

## 🚦 Health Monitoring

Each service provides health endpoints and participates in distributed health checking through Kafka messaging. The API Gateway aggregates health status from all services.

**Check overall system health:**
```bash
curl http://localhost:8080/api/v1/services/health
```

## 🔒 Security Considerations

- API Gateway acts as a security boundary
- Services communicate internally through Kafka
- Ready for service mesh integration (Istio/Linkerd)
- Environment-based configuration for secrets

## 🔧 Configuration

Services use environment variables for configuration:

- `PORT`: Service port (defaults provided)
- `KAFKA_BROKER`: Kafka broker address (default: localhost:9092)

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](docs/CONTRIBUTING.md) for details.

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🏷️ Version

Current version: `v1.0.0`

## 📞 Support

For questions and support:
- Create an issue in this repository
- Check the [Troubleshooting Guide](docs/TROUBLESHOOTING.md)
- Review the [Development Guide](docs/DEVELOPMENT.md)
