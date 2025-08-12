FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy shared module first
COPY shared/ ./shared/

# Copy service files
COPY services/catalog-service/go.mod services/catalog-service/go.sum ./services/catalog-service/
WORKDIR /app/services/catalog-service
RUN go mod download

# Copy the rest of the service code
COPY services/catalog-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/services/catalog-service/main .
EXPOSE 8082
CMD ["./main"]
