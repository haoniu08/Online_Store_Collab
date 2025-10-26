# Dockerfile for API Server (Order Receiver)
# This handles /orders/sync and /orders/async endpoints

# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the API server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o api-server ./cmd/server

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS (needed for AWS SDK)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/api-server .

# Expose port
EXPOSE 8080

# Run the API server
CMD ["./api-server"]