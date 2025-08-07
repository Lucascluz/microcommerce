FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy shared module first
COPY shared/ ./shared/

# Copy service files
COPY services/sales-service/go.mod services/sales-service/go.sum ./services/sales-service/
WORKDIR /app/services/sales-service
RUN go mod download

# Copy the rest of the service code
COPY services/sales-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/services/sales-service/main .
EXPOSE 8086
CMD ["./main"]
