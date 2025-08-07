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
warn('ℹ️ Open {tiltfile_path} in your favorite editor to get started.'.format(
    tiltfile_path=config.main_path))


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
    'payment-service',
    '.',  # Build from project root to include shared module
    dockerfile='./services/payment-service/Dockerfile'
)

docker_build(
    'product-service',
    '.',  # Build from project root to include shared module
    dockerfile='./services/product-service/Dockerfile'
)

docker_build(
    'user-service',
    '.',  # Build from project root to include shared module
    dockerfile='./services/user-service/Dockerfile'
)

# Deploy Kafka (using KRaft mode, no ZooKeeper needed)
k8s_yaml('./k8s/kafka/kafka.yaml')

# Deploy services
k8s_yaml('./k8s/payment-service/deployment.yaml')
k8s_yaml('./k8s/payment-service/service.yaml')
k8s_yaml('./k8s/product-service/deployment.yaml')
k8s_yaml('./k8s/product-service/service.yaml')
k8s_yaml('./k8s/user-service/deployment.yaml')
k8s_yaml('./k8s/user-service/service.yaml')

# Deploy API Gateway last (depends on other services)
k8s_yaml('./k8s/api-gateway/deployment.yaml')
k8s_yaml('./k8s/api-gateway/service.yaml')

# Port forwards for development
k8s_resource('api-gateway', port_forwards=['8080:8080'])
k8s_resource('payment-service', port_forwards=['8081:8081'])
k8s_resource('product-service', port_forwards=['8082:8082'])
k8s_resource('user-service', port_forwards=['8083:8083'])
k8s_resource('kafka', port_forwards=['9092:9092'])

# Resource dependencies - services depend on Kafka
k8s_resource('api-gateway', resource_deps=['kafka'])
k8s_resource('payment-service', resource_deps=['kafka'])
k8s_resource('product-service', resource_deps=['kafka'])
k8s_resource('user-service', resource_deps=['kafka'])


