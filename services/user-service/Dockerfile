FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy shared module first
COPY shared/ ./shared/

# Copy service files
COPY services/user-service/go.mod services/user-service/go.sum ./services/user-service/
WORKDIR /app/services/user-service
RUN go mod download

# Copy the rest of the service code
COPY services/user-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/services/user-service/main .
EXPOSE 8083
CMD ["./main"]
