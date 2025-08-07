# Deployment Guide

## Table of Contents

1. [Deployment Overview](#deployment-overview)
2. [Local Development Deployment](#local-development-deployment)
3. [Staging Environment](#staging-environment)
4. [Production Deployment](#production-deployment)
5. [Cloud Platforms](#cloud-platforms)
6. [CI/CD Pipeline](#cicd-pipeline)
7. [Monitoring and Logging](#monitoring-and-logging)
8. [Security Considerations](#security-considerations)
9. [Scaling Strategies](#scaling-strategies)
10. [Troubleshooting](#troubleshooting)

## Deployment Overview

MicroCommerce supports multiple deployment strategies:

- **Local Development**: Native Go, Docker, or Kubernetes with Tilt
- **Staging**: Kubernetes cluster with staging configurations
- **Production**: Production-grade Kubernetes with proper scaling and monitoring

### Deployment Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Development   │    │     Staging     │    │   Production    │
│                 │    │                 │    │                 │
│ • Local K8s     │    │ • Cloud K8s     │    │ • HA Cluster    │
│ • Tilt          │    │ • Auto-deploy   │    │ • Load Balancer │
│ • Hot Reload    │    │ • Integration   │    │ • Auto-scaling  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Local Development Deployment

### Option 1: Native Go (Fastest for Development)

```bash
# Start Kafka first (if not using external Kafka)
docker run -d --name kafka \
  -p 9092:9092 \
  -e KAFKA_NODE_ID=1 \
  -e KAFKA_PROCESS_ROLES=broker,controller \
  -e KAFKA_LISTENERS=PLAINTEXT://0.0.0.0:9092,CONTROLLER://0.0.0.0:9093 \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  -e KAFKA_CONTROLLER_LISTENER_NAMES=CONTROLLER \
  -e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT \
  -e KAFKA_CONTROLLER_QUORUM_VOTERS=1@localhost:9093 \
  -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
  apache/kafka:3.7.0

# Start all services
./scripts/run-all.sh
```

### Option 2: Kubernetes with Tilt (Recommended)

```bash
# Prerequisites
minikube start --cpus=4 --memory=8192

# Deploy with Tilt
tilt up

# Access services
kubectl port-forward svc/api-gateway 8080:8080
```

### Option 3: Docker Compose (Planned)

```yaml
# docker-compose.yml
version: '3.8'
services:
  kafka:
    image: apache/kafka:3.7.0
    ports:
      - "9092:9092"
    environment:
      KAFKA_NODE_ID: 1
      KAFKA_PROCESS_ROLES: broker,controller
      # ... other Kafka configs

  api-gateway:
    build:
      context: .
      dockerfile: services/api-gateway/Dockerfile
    ports:
      - "8080:8080"
    depends_on:
      - kafka
    environment:
      KAFKA_BROKER: kafka:9092

  # ... other services
```

## Staging Environment

### Kubernetes Staging Setup

1. **Prepare Kubernetes Cluster**

```bash
# For cloud providers (example with GKE)
gcloud container clusters create microcommerce-staging \
  --zone=us-central1-a \
  --num-nodes=3 \
  --machine-type=e2-standard-2

# Get credentials
gcloud container clusters get-credentials microcommerce-staging \
  --zone=us-central1-a
```

2. **Create Namespace**

```bash
kubectl create namespace microcommerce-staging
kubectl config set-context --current --namespace=microcommerce-staging
```

3. **Deploy with Helm (Planned)**

```bash
# Install Helm chart
helm install microcommerce-staging ./charts/microcommerce \
  --namespace microcommerce-staging \
  --values values-staging.yaml
```

4. **Manual Deployment**

```bash
# Deploy Kafka
kubectl apply -f k8s/kafka/

# Deploy services
kubectl apply -f k8s/api-gateway/
kubectl apply -f k8s/payment-service/
kubectl apply -f k8s/product-service/
kubectl apply -f k8s/user-service/
```

### Staging Configuration

```yaml
# values-staging.yaml
global:
  environment: staging
  domain: staging.microcommerce.example.com

api-gateway:
  replicas: 2
  resources:
    requests:
      memory: "256Mi"
      cpu: "250m"
    limits:
      memory: "512Mi"
      cpu: "500m"

kafka:
  replicas: 1
  persistence:
    enabled: true
    size: 10Gi
```

## Production Deployment

### Prerequisites

- **Kubernetes Cluster**: Version 1.25+
- **Container Registry**: Docker Hub, ECR, GCR, or ACR
- **Domain Name**: For external access
- **TLS Certificates**: Let's Encrypt or commercial certificates
- **Monitoring**: Prometheus and Grafana
- **Logging**: ELK stack or similar

### Production Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Load Balancer  │    │   Ingress      │    │   API Gateway   │
│                 │◄──►│   Controller    │◄──►│   (3 replicas)  │
│ • TLS Term.     │    │                 │    │                 │
│ • Rate Limiting │    │ • Routing       │    │ • Service Mesh  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                       │
                       ┌─────────────────────────────────┼─────────────────────────────────┐
                       │                                 │                                 │
                       ▼                                 ▼                                 ▼
                ┌─────────────┐                   ┌─────────────┐                   ┌─────────────┐
                │   Payment   │                   │   Product   │                   │    User     │
                │   Service   │                   │   Service   │                   │   Service   │
                │(3 replicas) │                   │(3 replicas) │                   │(2 replicas) │
                └─────────────┘                   └─────────────┘                   └─────────────┘
                       │                                 │                                 │
                       └─────────────────────────────────┼─────────────────────────────────┘
                                                         │
                ┌─────────────────────────────────────────┼─────────────────────────────────────────┐
                │                                         │                                         │
                ▼                                         ▼                                         ▼
         ┌─────────────┐                           ┌─────────────┐                           ┌─────────────┐
         │    Kafka    │                           │  Database   │                           │  Monitoring │
         │ (3 replicas)│                           │  Cluster    │                           │    Stack    │
         │             │                           │             │                           │             │
         └─────────────┘                           └─────────────┘                           └─────────────┘
```

### Production Deployment Steps

1. **Build and Push Images**

```bash
# Build images
docker build -t your-registry/api-gateway:v1.0.0 \
  -f services/api-gateway/Dockerfile .

docker build -t your-registry/payment-service:v1.0.0 \
  -f services/payment-service/Dockerfile .

docker build -t your-registry/product-service:v1.0.0 \
  -f services/product-service/Dockerfile .

docker build -t your-registry/user-service:v1.0.0 \
  -f services/user-service/Dockerfile .

# Push images
docker push your-registry/api-gateway:v1.0.0
docker push your-registry/payment-service:v1.0.0
docker push your-registry/product-service:v1.0.0
docker push your-registry/user-service:v1.0.0
```

2. **Create Production Namespace**

```bash
kubectl create namespace microcommerce-prod
kubectl config set-context --current --namespace=microcommerce-prod
```

3. **Deploy Infrastructure Components**

```bash
# Deploy Kafka cluster (production-ready)
kubectl apply -f k8s/kafka/kafka-production.yaml

# Deploy databases (if using in-cluster)
kubectl apply -f k8s/database/postgresql.yaml

# Deploy monitoring
kubectl apply -f k8s/monitoring/
```

4. **Deploy Application Services**

```bash
# Update image tags in production manifests
sed -i 's/latest/v1.0.0/g' k8s/*/deployment.yaml

# Deploy services
kubectl apply -f k8s/api-gateway/
kubectl apply -f k8s/payment-service/
kubectl apply -f k8s/product-service/
kubectl apply -f k8s/user-service/
```

5. **Configure Ingress**

```yaml
# k8s/ingress/production-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: microcommerce-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
spec:
  tls:
  - hosts:
    - api.microcommerce.example.com
    secretName: microcommerce-tls
  rules:
  - host: api.microcommerce.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-gateway
            port:
              number: 8080
```

### Production Configuration

```yaml
# k8s/api-gateway/deployment-production.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    spec:
      containers:
      - name: api-gateway
        image: your-registry/api-gateway:v1.0.0
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        env:
        - name: KAFKA_BROKER
          value: "kafka:9092"
        - name: PORT
          value: "8080"
```

## Cloud Platforms

### Amazon Web Services (AWS)

#### EKS Deployment

```bash
# Create EKS cluster
eksctl create cluster \
  --name microcommerce-prod \
  --version 1.25 \
  --region us-west-2 \
  --nodegroup-name workers \
  --node-type t3.medium \
  --nodes 3 \
  --nodes-min 1 \
  --nodes-max 10 \
  --managed

# Install ALB Ingress Controller
kubectl apply -f https://github.com/kubernetes-sigs/aws-load-balancer-controller/releases/download/v2.4.7/v2_4_7_full.yaml
```

#### AWS-specific Resources

```yaml
# Service with AWS Load Balancer
apiVersion: v1
kind: Service
metadata:
  name: api-gateway
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: nlb
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: api-gateway
```

### Google Cloud Platform (GCP)

#### GKE Deployment

```bash
# Create GKE cluster
gcloud container clusters create microcommerce-prod \
  --zone us-central1-a \
  --num-nodes 3 \
  --enable-autoscaling \
  --min-nodes 1 \
  --max-nodes 10 \
  --machine-type e2-standard-4
```

### Microsoft Azure

#### AKS Deployment

```bash
# Create resource group
az group create --name microcommerce-rg --location eastus

# Create AKS cluster
az aks create \
  --resource-group microcommerce-rg \
  --name microcommerce-prod \
  --node-count 3 \
  --enable-addons monitoring \
  --generate-ssh-keys
```

## CI/CD Pipeline

### GitHub Actions (Recommended)

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [main]
    tags: ['v*']

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.22
    - run: go test ./...

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2
    
    - name: Login to Registry
      uses: docker/login-action@v2
      with:
        registry: ${{ secrets.REGISTRY_URL }}
        username: ${{ secrets.REGISTRY_USERNAME }}
        password: ${{ secrets.REGISTRY_PASSWORD }}
    
    - name: Build and push images
      run: |
        docker build -t ${{ secrets.REGISTRY_URL }}/api-gateway:${{ github.sha }} \
          -f services/api-gateway/Dockerfile .
        docker push ${{ secrets.REGISTRY_URL }}/api-gateway:${{ github.sha }}
        
        # Build other services...

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: startsWith(github.ref, 'refs/tags/v')
    steps:
    - uses: actions/checkout@v3
    - name: Deploy to Kubernetes
      run: |
        # Update image tags
        sed -i 's/:latest/:${{ github.sha }}/g' k8s/*/deployment.yaml
        
        # Apply to cluster
        kubectl apply -f k8s/
```

### GitLab CI (Alternative)

```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - deploy

test:
  stage: test
  image: golang:1.22
  script:
    - go test ./...

build:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker build -t $CI_REGISTRY_IMAGE/api-gateway:$CI_COMMIT_SHA \
        -f services/api-gateway/Dockerfile .
    - docker push $CI_REGISTRY_IMAGE/api-gateway:$CI_COMMIT_SHA

deploy:
  stage: deploy
  image: bitnami/kubectl:latest
  script:
    - kubectl apply -f k8s/
  only:
    - tags
```

## Monitoring and Logging

### Prometheus and Grafana

```yaml
# k8s/monitoring/prometheus.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
    spec:
      containers:
      - name: prometheus
        image: prom/prometheus:latest
        ports:
        - containerPort: 9090
        volumeMounts:
        - name: config
          mountPath: /etc/prometheus
      volumes:
      - name: config
        configMap:
          name: prometheus-config
```

### Application Metrics

Add metrics to your Go services:

```go
// Add to main.go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    requestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
)

func init() {
    prometheus.MustRegister(requestsTotal)
}

func main() {
    router := gin.Default()
    
    // Metrics endpoint
    router.GET("/metrics", gin.WrapH(promhttp.Handler()))
    
    // ... rest of application
}
```

### Centralized Logging

```yaml
# k8s/logging/fluentd.yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: fluentd
spec:
  selector:
    matchLabels:
      name: fluentd
  template:
    metadata:
      labels:
        name: fluentd
    spec:
      containers:
      - name: fluentd
        image: fluent/fluentd-kubernetes-daemonset:v1-debian-elasticsearch
        env:
        - name: FLUENT_ELASTICSEARCH_HOST
          value: "elasticsearch"
        - name: FLUENT_ELASTICSEARCH_PORT
          value: "9200"
```

## Security Considerations

### Network Policies

```yaml
# k8s/security/network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: api-gateway-netpol
spec:
  podSelector:
    matchLabels:
      app: api-gateway
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: kafka
    ports:
    - protocol: TCP
      port: 9092
```

### Pod Security Standards

```yaml
# k8s/security/pod-security-policy.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: microcommerce-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
```

### Secrets Management

```bash
# Create secrets
kubectl create secret generic kafka-credentials \
  --from-literal=username=kafka-user \
  --from-literal=password=secure-password

kubectl create secret tls microcommerce-tls \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key
```

## Scaling Strategies

### Horizontal Pod Autoscaler

```yaml
# k8s/scaling/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-gateway-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-gateway
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### Vertical Pod Autoscaler

```yaml
# k8s/scaling/vpa.yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: api-gateway-vpa
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-gateway
  updatePolicy:
    updateMode: "Auto"
  resourcePolicy:
    containerPolicies:
    - containerName: api-gateway
      maxAllowed:
        cpu: 2
        memory: 4Gi
      minAllowed:
        cpu: 100m
        memory: 128Mi
```

### Cluster Autoscaler

```yaml
# For cloud providers, enable cluster autoscaling
# This automatically scales the number of nodes
```

## Troubleshooting

### Common Deployment Issues

#### Pod Stuck in Pending State

```bash
# Check node resources
kubectl describe nodes

# Check pod events
kubectl describe pod <pod-name>

# Check resource quotas
kubectl describe resourcequota
```

#### Service Not Accessible

```bash
# Check service endpoints
kubectl get endpoints <service-name>

# Check pod labels
kubectl get pods --show-labels

# Test internal connectivity
kubectl exec -it <pod-name> -- curl <service-name>:8080
```

#### Image Pull Errors

```bash
# Check image pull secrets
kubectl get secrets

# Verify image exists
docker pull <image-name>

# Check pod events
kubectl describe pod <pod-name>
```

### Performance Issues

#### High CPU/Memory Usage

```bash
# Check resource usage
kubectl top pods
kubectl top nodes

# Check metrics server
kubectl get apiservice v1beta1.metrics.k8s.io
```

#### Slow Response Times

```bash
# Check service latency
kubectl exec -it <pod-name> -- curl -w "@curl-format.txt" <service-url>

# Check Kafka lag
kubectl exec -it kafka-pod -- kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --describe --group <group-id>
```

### Rolling Back Deployments

```bash
# Check rollout history
kubectl rollout history deployment/api-gateway

# Rollback to previous version
kubectl rollout undo deployment/api-gateway

# Rollback to specific revision
kubectl rollout undo deployment/api-gateway --to-revision=2
```

### Health Check Debugging

```bash
# Check liveness/readiness probes
kubectl describe pod <pod-name>

# Test endpoints manually
kubectl exec -it <pod-name> -- curl localhost:8080/health

# Check probe configuration
kubectl get pod <pod-name> -o yaml | grep -A10 livenessProbe
```
