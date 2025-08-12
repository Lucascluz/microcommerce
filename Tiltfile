# Welcome to Tilt!
#   To get you started as quickly as possible, we have created a
#   starter Tiltfile for you.
#
#   Uncomment, modify, and delete any commands as needed for your
#   project's configuration.


# Output diagnostic messages
#   You can print log messages, warnings, and fatal errors, which will
#   appear in the (Tiltfile) resource in the web UI. Tiltfiles support
#   multiline strings and common string operations such as formatting.
#
#   More info: https://docs.tilt.dev/api.html#api.warn
print("""
-----------------------------------------------------------------
✨ Hello Tilt! This appears in the (Tiltfile) pane whenever Tilt
   evaluates this file.
-----------------------------------------------------------------
""".strip())


# Set Kubernetes context to minikube
allow_k8s_contexts(['minikube'])

# Build microservices with live update for faster development
docker_build(
    'api-gateway',
    '.',  # Build from project root to include shared module
    dockerfile='./services/api-gateway/Dockerfile',
    live_update=[
        sync('./services/api-gateway', '/app/services/api-gateway'),
        sync('./shared', '/app/shared'),
        run('cd /app/services/api-gateway && go build -o main cmd/main.go', trigger=['./services/api-gateway/**/*.go', './shared/**/*.go'])
    ]
)

docker_build(
    'catalog-service',
    '.',  # Build from project root to include shared module
    dockerfile='./services/catalog-service/Dockerfile'
)

docker_build(
    'transaction-service',
    '.',  # Build from project root to include shared module
    dockerfile='./services/transaction-service/Dockerfile'
)

docker_build(
    'user-service',
    '.',  # Build from project root to include shared module
    dockerfile='./services/user-service/Dockerfile'
)

docker_build(
    'notifications-service',
    '.',  # Build from project root to include shared module
    dockerfile='./services/notifications-service/Dockerfile'
)

docker_build(
    'visualization-service',
    '.',  # Build from project root to include shared module
    dockerfile='./services/visualization-service/Dockerfile'
)

# Deploy databases first
k8s_yaml('./k8s/database/postgres.yaml')
k8s_yaml('./k8s/database/redis.yaml')

# Deploy Kafka (using KRaft mode, no ZooKeeper needed)
k8s_yaml('./k8s/kafka/kafka.yaml')

# Deploy services
k8s_yaml('./k8s/catalog-service/deployment.yaml')
k8s_yaml('./k8s/catalog-service/service.yaml')
k8s_yaml('./k8s/transaction-service/deployment.yaml')
k8s_yaml('./k8s/transaction-service/service.yaml')
k8s_yaml('./k8s/user-service/deployment.yaml')
k8s_yaml('./k8s/user-service/service.yaml')
k8s_yaml('./k8s/notifications-service/deployment.yaml')
k8s_yaml('./k8s/notifications-service/service.yaml')
k8s_yaml('./k8s/visualization-service/deployment.yaml')
k8s_yaml('./k8s/visualization-service/service.yaml')

# Deploy API Gateway last (depends on other services)
k8s_yaml('./k8s/api-gateway/deployment.yaml')
k8s_yaml('./k8s/api-gateway/service.yaml')

# Port forwards for development
k8s_resource('api-gateway', port_forwards=['8080:8080'])
k8s_resource('catalog-service', port_forwards=['8082:8082'])
k8s_resource('transaction-service', port_forwards=['8081:8081'])
k8s_resource('user-service', port_forwards=['8083:8083'])
k8s_resource('notifications-service', port_forwards=['8087:8087'])
k8s_resource('visualization-service', port_forwards=['8089:8089'])
k8s_resource('kafka', port_forwards=['9092:9092'])
k8s_resource('postgres', port_forwards=['5432:5432'])
k8s_resource('redis', port_forwards=['6379:6379'])

# Resource dependencies - services depend on databases and Kafka
k8s_resource('api-gateway', resource_deps=['kafka', 'postgres', 'redis'])
k8s_resource('catalog-service', resource_deps=['kafka', 'postgres', 'redis'])
k8s_resource('transaction-service', resource_deps=['kafka', 'postgres', 'redis'])
k8s_resource('user-service', resource_deps=['kafka', 'postgres', 'redis'])
k8s_resource('notifications-service', resource_deps=['kafka', 'postgres', 'redis'])
k8s_resource('visualization-service', resource_deps=['kafka', 'postgres', 'redis'])


