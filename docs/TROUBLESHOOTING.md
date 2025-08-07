# Troubleshooting Guide

## Table of Contents

1. [Common Issues](#common-issues)
2. [Service-Specific Issues](#service-specific-issues)
3. [Kubernetes Issues](#kubernetes-issues)
4. [Kafka Issues](#kafka-issues)
5. [Networking Issues](#networking-issues)
6. [Performance Issues](#performance-issues)
7. [Development Issues](#development-issues)
8. [Debugging Tools](#debugging-tools)
9. [Logs Analysis](#logs-analysis)
10. [Getting Help](#getting-help)

## Common Issues

### Port Already in Use

**Problem**: Service fails to start with error "address already in use"

**Solution**:
```bash
# Find process using the port
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or kill all Go processes
pkill -f "go run"

# For specific service
pkill -f "main.go"
```

**Prevention**: Always stop services properly using `./scripts/stop-all.sh`

### Go Module Issues

**Problem**: `go mod` errors or dependency conflicts

**Solutions**:
```bash
# Clean module cache
go clean -modcache

# Refresh dependencies
go mod tidy
go mod download

# Verify module integrity
go mod verify

# Update specific dependency
go get -u github.com/gin-gonic/gin

# Fix replace directives
cd services/api-gateway && go mod edit -replace github.com/lucas/shared=../../shared
```

### Permission Denied Errors

**Problem**: Permission errors when running scripts or accessing files

**Solutions**:
```bash
# Make scripts executable
chmod +x scripts/run-all.sh
chmod +x scripts/stop-all.sh

# Fix ownership (if needed)
sudo chown -R $USER:$USER .

# Fix Docker permissions
sudo usermod -aG docker $USER
# Logout and login again
```

### Environment Variables Not Set

**Problem**: Services can't find configuration values

**Solutions**:
```bash
# Check current environment
env | grep KAFKA
env | grep PORT

# Set environment variables
export KAFKA_BROKER=localhost:9092
export PORT=8080

# Or create .env file
echo "KAFKA_BROKER=localhost:9092" > .env
echo "PORT=8080" >> .env

# Load .env file
source .env
```

## Service-Specific Issues

### API Gateway Issues

#### Gateway Returns 503 Service Unavailable

**Symptoms**: API Gateway starts but returns 503 for health checks

**Diagnosis**:
```bash
# Check if Kafka is running
docker ps | grep kafka
kubectl get pods | grep kafka

# Check API Gateway logs
kubectl logs -f deployment/api-gateway

# Test Kafka connectivity
kubectl exec -it api-gateway-pod -- telnet kafka 9092
```

**Solutions**:
1. Ensure Kafka is running and accessible
2. Check Kafka broker configuration
3. Verify network connectivity between services
4. Check firewall/security group settings

#### Health Check Timeout

**Symptoms**: Health check endpoint takes too long to respond

**Diagnosis**:
```bash
# Check service resources
kubectl top pods

# Check service logs
kubectl logs api-gateway-pod | grep -i timeout

# Test Kafka response time
time kubectl exec -it api-gateway-pod -- curl kafka:9092
```

**Solutions**:
1. Increase health check timeout values
2. Optimize Kafka consumer configuration
3. Add circuit breaker pattern
4. Scale Kafka if needed

### Payment Service Issues

#### Kafka Consumer Not Receiving Messages

**Symptoms**: Payment service starts but doesn't respond to ping messages

**Diagnosis**:
```bash
# Check consumer group
kubectl exec -it kafka-pod -- kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --describe --group payment-service-group

# Check topic messages
kubectl exec -it kafka-pod -- kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic service-ping --from-beginning

# Check service logs
kubectl logs -f payment-service-pod | grep -i kafka
```

**Solutions**:
1. Verify Kafka topic exists and has correct partitions
2. Check consumer group configuration
3. Restart consumer to reset offset
4. Verify Kafka broker health

### Product/User Service Issues

Similar patterns to Payment Service - check Kafka connectivity and consumer configuration.

## Kubernetes Issues

### Pod Stuck in Pending State

**Symptoms**: Pods remain in "Pending" status

**Diagnosis**:
```bash
# Check pod details
kubectl describe pod <pod-name>

# Check node resources
kubectl describe nodes

# Check resource quotas
kubectl describe resourcequota

# Check storage classes (if using PVCs)
kubectl get storageclass
```

**Common Causes & Solutions**:

1. **Insufficient Resources**:
   ```bash
   # Scale down other workloads or add nodes
   kubectl scale deployment unnecessary-app --replicas=0
   ```

2. **ImagePullBackOff**:
   ```bash
   # Check image exists
   docker pull <image-name>
   
   # Check image pull secrets
   kubectl get secrets
   ```

3. **Node Selector Issues**:
   ```bash
   # Check node labels
   kubectl get nodes --show-labels
   
   # Remove node selector if not needed
   kubectl patch deployment <name> -p '{"spec":{"template":{"spec":{"nodeSelector":null}}}}'
   ```

### Service Discovery Issues

**Symptoms**: Services can't reach each other

**Diagnosis**:
```bash
# Check service endpoints
kubectl get endpoints

# Test DNS resolution
kubectl exec -it <pod-name> -- nslookup kafka
kubectl exec -it <pod-name> -- nslookup api-gateway

# Check service configuration
kubectl get svc -o wide
```

**Solutions**:
1. Verify service selectors match pod labels
2. Check service ports and target ports
3. Ensure pods are running and ready
4. Verify network policies (if any)

### Persistent Volume Issues

**Symptoms**: Pods can't mount volumes or data is lost

**Diagnosis**:
```bash
# Check PV and PVC status
kubectl get pv,pvc

# Check storage class
kubectl describe storageclass

# Check pod events
kubectl describe pod <pod-name> | grep -A10 Events
```

**Solutions**:
1. Verify storage class supports the access mode
2. Check if PV has sufficient capacity
3. Ensure PVC is in the same namespace as the pod
4. Check storage backend health

## Kafka Issues

### Kafka Won't Start

**Symptoms**: Kafka pod fails to start or crashes repeatedly

**Diagnosis**:
```bash
# Check Kafka logs
kubectl logs kafka-pod

# Check Kafka configuration
kubectl describe configmap kafka-config

# Check storage
kubectl get pvc | grep kafka
```

**Common Issues & Solutions**:

1. **Insufficient Memory**:
   ```yaml
   # Increase memory limits
   resources:
     limits:
       memory: "2Gi"
     requests:
       memory: "1Gi"
   ```

2. **Storage Issues**:
   ```bash
   # Check if storage is available
   kubectl get pv
   kubectl describe pvc kafka-storage
   ```

3. **Configuration Errors**:
   ```bash
   # Validate Kafka configuration
   kubectl get configmap kafka-config -o yaml
   ```

### Topic Creation Issues

**Symptoms**: Services can't create or access topics

**Diagnosis**:
```bash
# List topics
kubectl exec -it kafka-pod -- kafka-topics.sh \
  --bootstrap-server localhost:9092 --list

# Describe specific topic
kubectl exec -it kafka-pod -- kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --describe --topic service-ping
```

**Solutions**:
```bash
# Manually create topic
kubectl exec -it kafka-pod -- kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create --topic service-ping \
  --partitions 3 --replication-factor 1

# Delete and recreate topic
kubectl exec -it kafka-pod -- kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --delete --topic service-ping
```

### Consumer Lag Issues

**Symptoms**: Messages are not being processed timely

**Diagnosis**:
```bash
# Check consumer group lag
kubectl exec -it kafka-pod -- kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --describe --group payment-service-group

# Check topic message count
kubectl exec -it kafka-pod -- kafka-run-class.sh \
  kafka.tools.GetOffsetShell \
  --broker-list localhost:9092 \
  --topic service-ping
```

**Solutions**:
1. Scale consumer instances
2. Optimize consumer processing logic
3. Increase partition count for better parallelism
4. Tune consumer configuration (fetch.min.bytes, etc.)

## Networking Issues

### Service-to-Service Communication Failures

**Symptoms**: Services can't communicate with each other

**Diagnosis**:
```bash
# Test connectivity from one pod to another
kubectl exec -it api-gateway-pod -- curl http://kafka:9092

# Check network policies
kubectl get networkpolicies

# Check service mesh configuration (if using Istio)
kubectl get destinationrules,virtualservices
```

**Solutions**:
1. Verify service names and ports
2. Check network policies allow traffic
3. Ensure services are in correct namespaces
4. Test with IP addresses to isolate DNS issues

### DNS Resolution Issues

**Symptoms**: "nslookup: can't resolve" errors

**Diagnosis**:
```bash
# Check DNS configuration
kubectl get configmap -n kube-system coredns -o yaml

# Test DNS from pod
kubectl exec -it <pod-name> -- nslookup kubernetes.default

# Check CoreDNS pods
kubectl get pods -n kube-system | grep coredns
```

**Solutions**:
1. Restart CoreDNS pods if needed
2. Check if DNS service is properly configured
3. Verify service exists in correct namespace
4. Use fully qualified domain names (FQDN) if needed

### Ingress Issues

**Symptoms**: External traffic can't reach services

**Diagnosis**:
```bash
# Check ingress status
kubectl get ingress

# Check ingress controller logs
kubectl logs -n ingress-nginx deployment/ingress-nginx-controller

# Verify ingress configuration
kubectl describe ingress microcommerce-ingress
```

**Solutions**:
1. Ensure ingress controller is installed and running
2. Check ingress annotations and configuration
3. Verify backend services are healthy
4. Check TLS certificate configuration

## Performance Issues

### High CPU Usage

**Symptoms**: Services are slow or unresponsive due to high CPU

**Diagnosis**:
```bash
# Check CPU usage
kubectl top pods
kubectl top nodes

# Profile application
kubectl port-forward pod/<pod-name> 6060:6060
go tool pprof http://localhost:6060/debug/pprof/profile
```

**Solutions**:
1. Optimize hot code paths
2. Add horizontal pod autoscaling
3. Implement caching where appropriate
4. Review and optimize algorithms

### High Memory Usage

**Symptoms**: Pods are killed due to OOM or high memory usage

**Diagnosis**:
```bash
# Check memory usage
kubectl top pods
kubectl describe pod <pod-name> | grep -A5 Events

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap
```

**Solutions**:
1. Identify memory leaks
2. Optimize data structures
3. Implement object pooling
4. Increase memory limits if necessary

### Slow Database Queries

**Symptoms**: API responses are slow due to database performance

**Diagnosis**:
```bash
# Check database metrics
kubectl exec -it postgres-pod -- psql -c "SELECT * FROM pg_stat_activity;"

# Analyze slow queries
kubectl logs database-pod | grep "slow query"
```

**Solutions**:
1. Add database indexes
2. Optimize query patterns
3. Implement query caching
4. Scale database (read replicas)

## Development Issues

### Hot Reload Not Working (Tilt)

**Symptoms**: Code changes don't trigger rebuilds

**Diagnosis**:
```bash
# Check Tilt logs
tilt logs <resource-name>

# Verify file watching
tilt get <resource-name>

# Check Tiltfile syntax
tilt doctor
```

**Solutions**:
1. Restart Tilt
2. Check file paths in live_update configuration
3. Verify Docker context includes changed files
4. Check file permissions

### Build Failures

**Symptoms**: Docker builds fail or produce incorrect images

**Diagnosis**:
```bash
# Check build context
docker build --no-cache -t test-build -f services/api-gateway/Dockerfile .

# Verify Dockerfile syntax
hadolint services/api-gateway/Dockerfile

# Check build logs
tilt logs <service-name>
```

**Solutions**:
1. Fix Dockerfile syntax errors
2. Ensure all required files are in build context
3. Check multi-stage build dependencies
4. Verify base image compatibility

### Go Build Issues

**Symptoms**: Go compilation fails

**Diagnosis**:
```bash
# Build locally
cd services/api-gateway && go build cmd/main.go

# Check for syntax errors
go vet ./...

# Verify dependencies
go mod verify
```

**Solutions**:
1. Fix Go syntax errors
2. Update incompatible dependencies
3. Resolve import conflicts
4. Check Go version compatibility

## Debugging Tools

### Essential Commands

```bash
# Kubernetes debugging
kubectl get pods -o wide
kubectl describe pod <pod-name>
kubectl logs -f <pod-name>
kubectl exec -it <pod-name> -- /bin/sh

# Docker debugging
docker ps
docker logs <container-id>
docker exec -it <container-id> /bin/sh

# Network debugging
kubectl exec -it <pod-name> -- netstat -tuln
kubectl exec -it <pod-name> -- curl -v <service-url>

# Kafka debugging
kubectl exec -it kafka-pod -- kafka-topics.sh --bootstrap-server localhost:9092 --list
kubectl exec -it kafka-pod -- kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic <topic>
```

### Useful Tools

1. **kubectx/kubens**: Switch between contexts and namespaces
2. **stern**: Multi-pod log tailing
3. **k9s**: Terminal UI for Kubernetes
4. **hey**: HTTP load testing
5. **kafkacat**: Kafka debugging tool

### Installation

```bash
# kubectx/kubens
curl -s https://raw.githubusercontent.com/ahmetb/kubectx/master/kubectx -o kubectx
curl -s https://raw.githubusercontent.com/ahmetb/kubectx/master/kubens -o kubens

# stern
brew install stern

# k9s
brew install derailed/k9s/k9s

# hey
go install github.com/rakyll/hey@latest
```

## Logs Analysis

### Structured Logging

Look for these patterns in logs:

```bash
# Error patterns
kubectl logs <pod-name> | grep -i "error\|fail\|exception"

# Performance patterns
kubectl logs <pod-name> | grep -i "slow\|timeout\|latency"

# Connection patterns
kubectl logs <pod-name> | grep -i "connect\|disconnect\|refused"

# Kafka patterns
kubectl logs <pod-name> | grep -i "kafka\|consumer\|producer"
```

### Log Aggregation

For production environments, consider:

1. **ELK Stack**: Elasticsearch, Logstash, Kibana
2. **Fluentd**: Log collection and forwarding
3. **Grafana Loki**: Log aggregation system
4. **Cloud Solutions**: AWS CloudWatch, GCP Logging, Azure Monitor

### Example Fluentd Configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
data:
  fluent.conf: |
    <source>
      @type tail
      path /var/log/containers/*microcommerce*.log
      pos_file /var/log/fluentd-containers.log.pos
      time_format %Y-%m-%dT%H:%M:%S.%NZ
      tag kubernetes.*
      format json
    </source>
    
    <match kubernetes.**>
      @type elasticsearch
      host elasticsearch
      port 9200
      index_name microcommerce
    </match>
```

## Getting Help

### Before Asking for Help

1. **Search documentation** and existing issues
2. **Try basic troubleshooting** steps
3. **Gather relevant information**:
   - Error messages and logs
   - Environment details (OS, versions)
   - Steps to reproduce
   - What you've already tried

### Information to Include

When reporting issues, include:

```bash
# System information
uname -a
go version
docker --version
kubectl version

# Kubernetes cluster info
kubectl cluster-info
kubectl get nodes

# Service status
kubectl get pods -o wide
kubectl get svc
kubectl get ing

# Recent logs
kubectl logs <pod-name> --tail=50

# Resource usage
kubectl top pods
kubectl top nodes
```

### Where to Get Help

1. **Documentation**: Check all docs in this repository
2. **GitHub Issues**: Search existing issues and create new ones
3. **GitHub Discussions**: For general questions
4. **Stack Overflow**: Tag with `microcommerce` and `go`
5. **Community Slack/Discord**: If available

### Creating Good Issue Reports

Use this template:

```markdown
## Problem Description
Describe what's not working as expected.

## Environment
- OS: [e.g., Ubuntu 20.04]
- Go version: [e.g., 1.22.0]
- Kubernetes version: [e.g., 1.25.0]
- Deployment method: [native/tilt/kubectl]

## Steps to Reproduce
1. Step one
2. Step two
3. Step three

## Expected Behavior
What should happen.

## Actual Behavior
What actually happens.

## Logs
```
Paste relevant logs here
```

## Additional Context
Any other relevant information.
```

### Emergency Contacts

For critical production issues:
- Check on-call rotation
- Escalate through proper channels
- Document incident for post-mortem

Remember: Most issues have been encountered before. Take time to search and understand the problem before escalating.
