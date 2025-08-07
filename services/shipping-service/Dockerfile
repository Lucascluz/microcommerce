FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy shared module first
COPY shared/ ./shared/

# Copy service files
COPY services/shipping-service/go.mod services/shipping-service/go.sum ./services/shipping-service/
WORKDIR /app/services/shipping-service
RUN go mod download

# Copy the rest of the service code
COPY services/shipping-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/services/shipping-service/main .
EXPOSE 8085
CMD ["./main"]
