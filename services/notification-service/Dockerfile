FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy shared module first
COPY shared/ ./shared/

# Copy service files
COPY services/notification-service/go.mod services/notification-service/go.sum ./services/notification-service/
WORKDIR /app/services/notification-service
RUN go mod download

# Copy the rest of the service code
COPY services/notification-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/services/notification-service/main .
EXPOSE 8087
CMD ["./main"]
