FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy shared module first
COPY shared/ ./shared/

# Copy service files
COPY services/payment-service/go.mod services/payment-service/go.sum ./services/payment-service/
WORKDIR /app/services/payment-service
RUN go mod download

# Copy the rest of the service code
COPY services/payment-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/services/payment-service/main .
EXPOSE 8081
CMD ["./main"]