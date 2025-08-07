FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy shared module first
COPY shared/ ./shared/

# Copy service files
COPY services/review-service/go.mod services/review-service/go.sum ./services/review-service/
WORKDIR /app/services/review-service
RUN go mod download

# Copy the rest of the service code
COPY services/review-service/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/services/review-service/main .
EXPOSE 8088
CMD ["./main"]
