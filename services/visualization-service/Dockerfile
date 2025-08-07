FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy shared module first
COPY shared/ ./shared/

# Copy service files
COPY services/visualization-service/go.mod services/visualization-service/go.sum ./services/visualization-service/
WORKDIR /app/services/visualization-service
RUN go mod download

# Copy the rest of the service code
COPY services/visualization-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/services/visualization-service/main .
EXPOSE 8089
CMD ["./main"]
