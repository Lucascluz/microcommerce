# API Documentation

## Overview

This document provides comprehensive information about the MicroCommerce API endpoints. The API follows RESTful principles and uses JSON for data exchange.

**Architecture Update:** The system has been restructured into 5 core services:
- **user-service** (8083): Authentication, profiles, account management
- **catalog-service** (8082): Product listings, reviews, inventory
- **transaction-service** (8081): Orders, payments, sales, shipping
- **notifications-service** (8087): Email, push notifications
- **visualization-service** (8089): Analytics, reports, observability

## Base URL

**Local Development:**
```
http://localhost:8080
```

**Production:** (To be configured)
```
https://api.microcommerce.example.com
```

## Shared Infrastructure

**Database:**
- PostgreSQL (shared): Port 5432 - Primary data storage
- Redis (shared): Port 6379 - Caching and sessions

**Message Queue:**
- Kafka: Port 9092 - Inter-service communication

## API Versioning

The API uses URI versioning with the pattern `/api/v{version}/`. Currently supported version: `v1`

Base API Path: `/api/v1`

## Authentication

**Current:** No authentication required (development phase)

**Planned:** JWT-based authentication with the following headers:
```http
Authorization: Bearer <jwt-token>
Content-Type: application/json
```

## Rate Limiting

**Current:** No rate limiting implemented

**Planned:** 
- 1000 requests per hour for authenticated users
- 100 requests per hour for unauthenticated users

Rate limit headers will be included in responses:
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## API Gateway Endpoints

### System Health

#### Get Overall System Health

Get the health status of all microservices in the system.

- **URL:** `/api/v1/services/health`
- **Method:** `GET`
- **Auth Required:** No

**Response:**

```json
{
  "services": [
    {
      "status": "healthy",
      "service": "payment-service",
      "timestamp": "2024-01-15T10:30:00Z"
    },
    {
      "status": "healthy", 
      "service": "product-service",
      "timestamp": "2024-01-15T10:30:00Z"
    },
    {
      "status": "healthy",
      "service": "user-service", 
      "timestamp": "2024-01-15T10:30:00Z"
    }
  ],
  "total_services": 3,
  "responding_services": 3
}
```

**Status Codes:**
- `200 OK` - All services healthy or partial services responding
- `503 Service Unavailable` - Cannot communicate with Kafka or critical failure

#### Gateway Health Check

Simple health check for the API Gateway itself.

- **URL:** `/`
- **Method:** `GET`
- **Auth Required:** No

**Response:**

```json
{
  "status": "healthy",
  "service": "api-gateway"
}
```

**Status Codes:**
- `200 OK` - Gateway is healthy

## Service-Specific Endpoints

### Payment Service

#### Payment Service Health

- **URL:** `http://localhost:8081/`
- **Method:** `GET`
- **Auth Required:** No

**Response:**

```json
{
  "status": "healthy",
  "service": "payment-service"
}
```

**Planned Payment Endpoints:**

#### Process Payment

- **URL:** `/api/v1/payments`
- **Method:** `POST`
- **Auth Required:** Yes (planned)

**Request Body:**
```json
{
  "amount": 99.99,
  "currency": "USD",
  "payment_method": {
    "type": "card",
    "card_number": "4111111111111111",
    "expiry_month": 12,
    "expiry_year": 2025,
    "cvc": "123"
  },
  "order_id": "ord_123456"
}
```

**Response:**
```json
{
  "payment_id": "pay_789012",
  "status": "succeeded",
  "amount": 99.99,
  "currency": "USD", 
  "order_id": "ord_123456",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Product Service

#### Product Service Health

- **URL:** `http://localhost:8082/`
- **Method:** `GET`
- **Auth Required:** No

**Response:**

```json
{
  "status": "healthy",
  "service": "product-service"
}
```

**Planned Product Endpoints:**

#### Get Products

- **URL:** `/api/v1/products`
- **Method:** `GET`
- **Auth Required:** No

**Query Parameters:**
- `page` (integer): Page number (default: 1)
- `limit` (integer): Items per page (default: 20, max: 100)
- `category` (string): Filter by category
- `search` (string): Search in product name and description
- `sort` (string): Sort by field (name, price, created_at)
- `order` (string): Sort order (asc, desc)

**Response:**
```json
{
  "products": [
    {
      "id": "prod_123",
      "name": "Wireless Headphones",
      "description": "High-quality wireless headphones with noise cancellation",
      "price": 299.99,
      "currency": "USD",
      "category": "electronics",
      "stock_quantity": 50,
      "images": [
        "https://example.com/images/headphones1.jpg"
      ],
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

#### Get Product by ID

- **URL:** `/api/v1/products/{id}`
- **Method:** `GET`
- **Auth Required:** No

**Response:**
```json
{
  "id": "prod_123",
  "name": "Wireless Headphones",
  "description": "High-quality wireless headphones with noise cancellation",
  "price": 299.99,
  "currency": "USD",
  "category": "electronics",
  "stock_quantity": 50,
  "images": [
    "https://example.com/images/headphones1.jpg"
  ],
  "specifications": {
    "battery_life": "30 hours",
    "weight": "250g",
    "connectivity": "Bluetooth 5.0"
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

#### Create Product

- **URL:** `/api/v1/products`
- **Method:** `POST`
- **Auth Required:** Yes (Admin role)

**Request Body:**
```json
{
  "name": "Wireless Headphones",
  "description": "High-quality wireless headphones with noise cancellation",
  "price": 299.99,
  "currency": "USD",
  "category": "electronics",
  "stock_quantity": 50,
  "images": [
    "https://example.com/images/headphones1.jpg"
  ],
  "specifications": {
    "battery_life": "30 hours",
    "weight": "250g",
    "connectivity": "Bluetooth 5.0"
  }
}
```

### User Service

#### User Service Health

- **URL:** `http://localhost:8083/`
- **Method:** `GET`
- **Auth Required:** No

**Response:**

```json
{
  "status": "healthy",
  "service": "user-service"
}
```

**Planned User Endpoints:**

#### User Registration

- **URL:** `/api/v1/users/register`
- **Method:** `POST`
- **Auth Required:** No

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890"
}
```

**Response:**
```json
{
  "user_id": "usr_123456",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "created_at": "2024-01-15T10:30:00Z",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "ref_789012345"
}
```

#### User Login

- **URL:** `/api/v1/users/login`
- **Method:** `POST`
- **Auth Required:** No

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**Response:**
```json
{
  "user_id": "usr_123456",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "ref_789012345"
}
```

#### Get User Profile

- **URL:** `/api/v1/users/profile`
- **Method:** `GET`
- **Auth Required:** Yes

**Response:**
```json
{
  "user_id": "usr_123456",
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "address": {
    "street": "123 Main St",
    "city": "New York",
    "state": "NY",
    "zip": "10001",
    "country": "US"
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

## Error Handling

The API uses conventional HTTP response codes to indicate success or failure.

### Error Response Format

All error responses follow this format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "The request contains invalid parameters",
    "details": {
      "field": "email",
      "reason": "Invalid email format"
    },
    "request_id": "req_123456789"
  }
}
```

### HTTP Status Codes

- `200 OK` - Request succeeded
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource already exists
- `422 Unprocessable Entity` - Validation error
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - Service temporarily unavailable

### Common Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Request validation failed |
| `AUTHENTICATION_REQUIRED` | Valid authentication required |
| `AUTHORIZATION_FAILED` | Insufficient permissions |
| `RESOURCE_NOT_FOUND` | Requested resource does not exist |
| `RESOURCE_CONFLICT` | Resource already exists |
| `RATE_LIMIT_EXCEEDED` | Too many requests |
| `SERVICE_UNAVAILABLE` | Service temporarily unavailable |
| `INTERNAL_ERROR` | Unexpected server error |

## Request/Response Examples

### Successful Request

**Request:**
```http
GET /api/v1/services/health HTTP/1.1
Host: localhost:8080
Content-Type: application/json
```

**Response:**
```http
HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 234

{
  "services": [
    {
      "status": "healthy",
      "service": "payment-service",
      "timestamp": "2024-01-15T10:30:00Z"
    }
  ],
  "total_services": 3,
  "responding_services": 1
}
```

### Error Request

**Request:**
```http
GET /api/v1/nonexistent HTTP/1.1
Host: localhost:8080
Content-Type: application/json
```

**Response:**
```http
HTTP/1.1 404 Not Found
Content-Type: application/json
Content-Length: 145

{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "The requested endpoint does not exist",
    "request_id": "req_123456789"
  }
}
```

## Webhooks (Planned)

Future webhook implementation for real-time notifications:

### Payment Webhooks

**Endpoint:** `POST /webhooks/payments`

**Events:**
- `payment.succeeded`
- `payment.failed`
- `payment.refunded`

### Order Webhooks

**Endpoint:** `POST /webhooks/orders`

**Events:**
- `order.created`
- `order.updated`
- `order.cancelled`
- `order.fulfilled`

## SDK and Client Libraries (Planned)

Future SDK availability:
- **JavaScript/TypeScript**: npm package
- **Python**: PyPI package
- **Go**: Go module
- **Java**: Maven package

## Postman Collection

A Postman collection with all endpoints will be available at:
`docs/postman/MicroCommerce.postman_collection.json`

## OpenAPI Specification

OpenAPI 3.0 specification will be available at:
- **Local:** `http://localhost:8080/api/docs`
- **File:** `docs/openapi.yaml`

## Testing

### Test Environment

**Base URL:** `http://localhost:8080` (local development)

### Test Data

Test user credentials:
```json
{
  "email": "test@example.com",
  "password": "testPassword123"
}
```

Test payment card:
```json
{
  "card_number": "4111111111111111",
  "expiry_month": 12,
  "expiry_year": 2025,
  "cvc": "123"
}
```

## Changelog

### v1.0.0 (Current)
- Initial API documentation
- Health check endpoints
- Basic service structure

### Planned v1.1.0
- User authentication endpoints
- Basic product management
- Payment processing

### Planned v1.2.0
- Order management
- Advanced product features
- Webhook support
