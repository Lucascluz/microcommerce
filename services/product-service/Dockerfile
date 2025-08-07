FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy shared module first
COPY shared/ ./shared/

# Copy service files
COPY services/product-service/go.mod services/product-service/go.sum ./services/product-service/
WORKDIR /app/services/product-service
RUN go mod download

# Copy the rest of the service code
COPY services/product-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/services/product-service/main .
EXPOSE 8082
CMD ["./main"]